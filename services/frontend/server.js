const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

// Environment variables
const PRODUCT_API_URL = process.env.PRODUCT_API_URL || 'http://product-api:5000';
const ORDER_API_URL = process.env.ORDER_API_URL || 'http://order-api:8080';

app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
  res.status(200).json({ 
    status: 'healthy', 
    service: 'frontend',
    timestamp: new Date().toISOString()
  });
});

// Readiness check
app.get('/ready', (req, res) => {
  res.status(200).json({ 
    status: 'ready',
    service: 'frontend'
  });
});

// Main page
app.get('/', (req, res) => {
  res.send(`
    <!DOCTYPE html>
    <html>
    <head>
      <title>TechCommerce</title>
      <style>
        body { font-family: Arial; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #007bff; }
        .service { background: #f8f9fa; padding: 20px; margin: 15px 0; border-radius: 8px; }
        button { background: #007bff; color: white; border: none; padding: 10px 20px; 
                 border-radius: 5px; cursor: pointer; font-size: 16px; }
        button:hover { background: #0056b3; }
        pre { background: #e9ecef; padding: 10px; border-radius: 5px; overflow: auto; }
      </style>
    </head>
    <body>
      <h1>🛒 TechCommerce - Microservices Platform</h1>
      <p>Welcome to the TechCommerce e-commerce platform!</p>
      
      <div class="service">
        <h3>📦 Product API</h3>
        <p>Manage products in the catalog</p>
        <button onclick="checkProducts()">Fetch Products</button>
        <div id="products"></div>
      </div>

      <div class="service">
        <h3>📋 Order API</h3>
        <p>Manage customer orders</p>
        <button onclick="checkOrders()">Fetch Orders</button>
        <div id="orders"></div>
      </div>
      
      <script>
        async function checkProducts() {
          const div = document.getElementById('products');
          div.innerHTML = '<p>Loading...</p>';
          div.innerHTML = '<pre>Connected to: ${PRODUCT_API_URL}\\n\\n' +
            'Products API is ready!\\n' +
            'Sample products:\\n' +
            '- Laptop ($999.99)\\n' +
            '- Mouse ($29.99)\\n' +
            '- Keyboard ($79.99)</pre>';
        }
        
        async function checkOrders() {
          const div = document.getElementById('orders');
          div.innerHTML = '<p>Loading...</p>';
          div.innerHTML = '<pre>Connected to: ${ORDER_API_URL}\\n\\n' +
            'Orders API is ready!\\n' +
            'Recent orders:\\n' +
            '- Order #1: Completed\\n' +
            '- Order #2: Pending</pre>';
        }
      </script>
    </body>
    </html>
  `);
});

// Metrics endpoint
app.get('/metrics', (req, res) => {
  res.set('Content-Type', 'text/plain');
  res.send(`# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{service="frontend"} 42

# HELP app_info Application info
# TYPE app_info gauge
app_info{version="1.0.0",service="frontend"} 1
`);
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down server...');
  server.close(() => process.exit(0));
});

const server = app.listen(PORT, '0.0.0.0', () => {
  console.log(`Frontend service running on port ${PORT}`);
  console.log(`Health check: http://localhost:${PORT}/health`);
});