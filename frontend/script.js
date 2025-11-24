// API endpoints - In production, these would be environment variables
const API_BASE = '';    // empty because frontend is served from the same origin using a reverse proxy
const PRODUCTS_API = `${API_BASE}/api/products`;
const INVENTORY_API = `${API_BASE}/api/inventory`;
const ORDERS_API = `${API_BASE}/api/orders`;

// Global state
let products = [];
let inventory = [];
let orders = [];

// DOM elements
const navButtons = document.querySelectorAll('.nav-btn');
const sections = document.querySelectorAll('.section');
const messageDiv = document.getElementById('message');

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    setupNavigation();
    loadProducts();
    loadInventory();
    loadOrders();
    setupOrderForm();
});

// Navigation
function setupNavigation() {
    navButtons.forEach(button => {
        button.addEventListener('click', function() {
            const sectionId = this.dataset.section;
            switchSection(sectionId);
        });
    });
}

function switchSection(sectionId) {
    // Update nav buttons
    navButtons.forEach(btn => btn.classList.remove('active'));
    document.querySelector(`[data-section="${sectionId}"]`).classList.add('active');
    
    // Update sections
    sections.forEach(section => section.classList.remove('active'));
    document.getElementById(sectionId).classList.add('active');
    
    // Refresh data when switching to certain sections
    if (sectionId === 'inventory') {
        loadInventory();
    } else if (sectionId === 'orders') {
        loadOrders();
    }
}

// API calls
async function fetchData(url) {
    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error('Fetch error:', error);
        showMessage('Error fetching data: ' + error.message, 'error');
        return null;
    }
}

async function postData(url, data) {
    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data)
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error('Post error:', error);
        showMessage('Error: ' + error.message, 'error');
        return null;
    }
}

// Load data functions
async function loadProducts() {
    const productsContainer = document.getElementById('products-list');
    productsContainer.innerHTML = '<div class="loading">Loading products...</div>';
    
    const data = await fetchData(`${PRODUCTS_API}/products`);
    if (data) {
        products = data;
        displayProducts(data);
        updateProductSelects();
    }
}

async function loadInventory() {
    const inventoryContainer = document.getElementById('inventory-list');
    inventoryContainer.innerHTML = '<div class="loading">Loading inventory...</div>';
    
    const data = await fetchData(`${INVENTORY_API}/inventory`);
    if (data) {
        inventory = data;
        displayInventory(data);
    }
}

async function loadOrders() {
    const ordersContainer = document.getElementById('orders-list');
    ordersContainer.innerHTML = '<div class="loading">Loading orders...</div>';
    
    const data = await fetchData(`${ORDERS_API}/orders`);
    if (data) {
        orders = data;
        displayOrders(data);
    }
}

// Display functions
function displayProducts(products) {
    const container = document.getElementById('products-list');
    
    if (!products || products.length === 0) {
        container.innerHTML = '<div class="loading">No products found.</div>';
        return;
    }
    
    container.innerHTML = products.map(product => `
        <div class="item-card">
            <h3>${product.name}</h3>
            <p><strong>ID:</strong> ${product.id}</p>
            <p><strong>Category:</strong> ${product.category}</p>
            <p><strong>Description:</strong> ${product.description}</p>
            <p class="price">$${product.price.toFixed(2)}</p>
        </div>
    `).join('');
}

function displayInventory(inventory) {
    const container = document.getElementById('inventory-list');
    
    if (!inventory || inventory.length === 0) {
        container.innerHTML = '<div class="loading">No inventory data found.</div>';
        return;
    }
    
    container.innerHTML = inventory.map(item => {
        const product = products.find(p => p.id === item.product_id);
        const productName = product ? product.name : `Product ${item.product_id}`;
        const available = item.stock - item.reserved;
        
        return `
            <div class="item-card">
                <h3>${productName}</h3>
                <p><strong>Product ID:</strong> ${item.product_id}</p>
                <div class="stock-info">
                    <span class="stock-available">Available: ${available}</span>
                    <span class="stock-reserved">Reserved: ${item.reserved}</span>
                </div>
                <p><strong>Total Stock:</strong> ${item.stock}</p>
            </div>
        `;
    }).join('');
}

function displayOrders(orders) {
    const container = document.getElementById('orders-list');
    
    if (!orders || orders.length === 0) {
        container.innerHTML = '<div class="loading">No orders found.</div>';
        return;
    }
    
    container.innerHTML = orders.map(order => `
        <div class="item-card">
            <h3>Order #${order.id}</h3>
            <p><strong>Customer ID:</strong> ${order.customer_id}</p>
            <p><strong>Status:</strong> <span class="status ${order.status}">${order.status}</span></p>
            <p><strong>Total:</strong> $${order.total_amount.toFixed(2)}</p>
            <p><strong>Created:</strong> ${new Date(order.created_at).toLocaleDateString()}</p>
            <div><strong>Items:</strong></div>
            ${order.items.map(item => {
                const product = products.find(p => p.id === item.product_id);
                const productName = product ? product.name : `Product ${item.product_id}`;
                return `<p>• ${productName} x${item.quantity}</p>`;
            }).join('')}
            ${order.status === 'pending' ? `
                <button onclick="fulfillOrder(${order.id})" style="margin-top: 10px; padding: 5px 10px; background: #27ae60; color: white; border: none; border-radius: 3px; cursor: pointer;">
                    Fulfill Order
                </button>
            ` : ''}
        </div>
    `).join('');
}

// Order form functions
function setupOrderForm() {
    const form = document.getElementById('order-form');
    const addItemBtn = document.getElementById('add-item');
    
    form.addEventListener('submit', handleOrderSubmit);
    addItemBtn.addEventListener('click', addOrderItem);
    
    // Setup remove buttons for initial item
    setupRemoveButtons();
}

function updateProductSelects() {
    const selects = document.querySelectorAll('.product-select');
    const optionsHTML = '<option value="">Select Product</option>' + 
        products.map(product => `<option value="${product.id}">${product.name} - $${product.price.toFixed(2)}</option>`).join('');
    
    selects.forEach(select => {
        select.innerHTML = optionsHTML;
    });
}

function addOrderItem() {
    const container = document.getElementById('order-items');
    const newItem = document.createElement('div');
    newItem.className = 'order-item';
    newItem.innerHTML = `
        <select class="product-select" required>
            <option value="">Select Product</option>
            ${products.map(product => `<option value="${product.id}">${product.name} - $${product.price.toFixed(2)}</option>`).join('')}
        </select>
        <input type="number" class="quantity-input" placeholder="Quantity" min="1" required>
        <button type="button" class="remove-item">Remove</button>
    `;
    
    container.appendChild(newItem);
    setupRemoveButtons();
}

function setupRemoveButtons() {
    const removeButtons = document.querySelectorAll('.remove-item');
    removeButtons.forEach(button => {
        button.onclick = function() {
            const orderItems = document.querySelectorAll('.order-item');
            if (orderItems.length > 1) {
                this.parentElement.remove();
            } else {
                showMessage('At least one item is required', 'error');
            }
        };
    });
}

async function handleOrderSubmit(e) {
    e.preventDefault();
    
    const customerId = parseInt(document.getElementById('customer-id').value);
    const orderItems = document.querySelectorAll('.order-item');
    
    // Note: The backend inventory service expects { stock: ... } for reservation,
    // but the orders service handles this mapping. The frontend should keep using 'quantity'.
    const items = [];
    for (let item of orderItems) {
        const productId = parseInt(item.querySelector('.product-select').value);
        const quantity = parseInt(item.querySelector('.quantity-input').value);
        
        if (productId && quantity > 0) {
            items.push({
                product_id: productId,
                quantity: quantity
            });
        }
    }
    
    if (items.length === 0) {
        showMessage('Please add at least one item to the order', 'error');
        return;
    }
    
    const orderData = {
        customer_id: customerId,
        items: items
    };
    
    const result = await postData(`${ORDERS_API}/orders`, orderData);
    if (result) {
        showMessage('Order created successfully!', 'success');
        document.getElementById('order-form').reset();
        loadOrders();
        loadInventory(); // Refresh inventory to show reserved items
    }
}

async function fulfillOrder(orderId) {
    const result = await postData(`${ORDERS_API}/orders/${orderId}/fulfill`, {});
    if (result) {
        showMessage('Order fulfilled successfully!', 'success');
        loadOrders();
        loadInventory(); // Refresh inventory to show updated stock
    }
}

// Utility functions
function showMessage(text, type) {
    messageDiv.textContent = text;
    messageDiv.className = `message ${type} show`;
    
    setTimeout(() => {
        messageDiv.classList.remove('show');
    }, 3000);
}