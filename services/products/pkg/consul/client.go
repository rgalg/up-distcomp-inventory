package consul

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

type Client struct {
	client      *api.Client
	serviceName string
	servicePort int
}

// NewClient creates a new Consul client
func NewClient() (*Client, error) {
	consulHost := os.Getenv("CONSUL_HOST")
	if consulHost == "" {
		consulHost = "localhost"
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		return nil, fmt.Errorf("SERVICE_NAME environment variable is required")
	}

	servicePortStr := os.Getenv("SERVICE_PORT")
	if servicePortStr == "" {
		return nil, fmt.Errorf("SERVICE_PORT environment variable is required")
	}

	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SERVICE_PORT: %v", err)
	}

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:8500", consulHost)

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %v", err)
	}

	return &Client{
		client:      client,
		serviceName: serviceName,
		servicePort: servicePort,
	}, nil
}

// RegisterService registers the service with Consul
func (c *Client) RegisterService() error {
	// Get container hostname for service registration
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %v", err)
	}

	registration := &api.AgentServiceRegistration{
		ID:   fmt.Sprintf("%s-%s", c.serviceName, hostname),
		Name: c.serviceName,
		Port: c.servicePort,
		Address: hostname, // Use container hostname for internal Docker networking
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", hostname, c.servicePort),
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err = c.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}

	log.Printf("Service %s registered with Consul at %s:%d", c.serviceName, hostname, c.servicePort)
	return nil
}

// DeregisterService removes the service from Consul
func (c *Client) DeregisterService() error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %v", err)
	}

	serviceID := fmt.Sprintf("%s-%s", c.serviceName, hostname)
	err = c.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %v", err)
	}

	log.Printf("Service %s deregistered from Consul", c.serviceName)
	return nil
}

// DiscoverService finds a healthy service instance
func (c *Client) DiscoverService(serviceName string) (string, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("failed to discover service %s: %v", serviceName, err)
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances of service %s found", serviceName)
	}

	// Return the first healthy service instance
	service := services[0]
	endpoint := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)
	
	log.Printf("Discovered service %s at %s", serviceName, endpoint)
	return endpoint, nil
}

// WaitForConsul waits for Consul to be available
func (c *Client) WaitForConsul(maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		_, err := c.client.Status().Leader()
		if err == nil {
			log.Printf("Consul is available")
			return nil
		}
		
		log.Printf("Waiting for Consul to be available... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
	
	return fmt.Errorf("consul not available after %d retries", maxRetries)
}