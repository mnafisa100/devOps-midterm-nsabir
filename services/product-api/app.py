from flask import Flask, jsonify, request
from flask_cors import CORS
import os
import logging
import time
import random
from datetime import datetime

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app)

# Configuration
PORT = int(os.getenv('PORT', 5000))
ENV = os.getenv('ENVIRONMENT', 'development')

# In-memory database
products = [
    {"id": 1, "name": "Laptop", "price": 999.99, "stock": 50},
    {"id": 2, "name": "Mouse", "price": 29.99, "stock": 200},
    {"id": 3, "name": "Keyboard", "price": 79.99, "stock": 150},
    {"id": 4, "name": "Monitor", "price": 299.99, "stock": 75},
    {"id": 5, "name": "Webcam", "price": 89.99, "stock": 100},
]

app_state = {
    'startup_time': datetime.now(),
    'ready': True
}


@app.route('/health', methods=['GET'])
def health():
    """Liveness probe"""
    return jsonify({
        'status': 'healthy',
        'service': 'product-api',
        'timestamp': datetime.now().isoformat()
    }), 200


@app.route('/ready', methods=['GET'])
def ready():
    """Readiness probe"""
    if not app_state['ready']:
        return jsonify({'status': 'not ready'}), 503

    return jsonify({
        'status': 'ready',
        'service': 'product-api'
    }), 200


@app.route('/api/products', methods=['GET'])
def get_products():
    """Get all products"""
    time.sleep(random.uniform(0.01, 0.05))  # Simulate processing

    logger.info(f"Fetching all products - Total: {len(products)}")
    return jsonify({
        'success': True,
        'count': len(products),
        'products': products
    }), 200


@app.route('/api/products/<int:product_id>', methods=['GET'])
def get_product(product_id):
    """Get specific product"""
    product = next((p for p in products if p['id'] == product_id), None)

    if product:
        return jsonify({'success': True, 'product': product}), 200
    else:
        return jsonify({'success': False, 'error': 'Product not found'}), 404


@app.route('/api/products', methods=['POST'])
def create_product():
    """Create new product"""
    data = request.get_json()

    required = ['name', 'price', 'stock']
    if not all(field in data for field in required):
        return jsonify({'success': False, 'error': 'Missing required fields'}), 400

    new_product = {
        'id': max([p['id'] for p in products]) + 1 if products else 1,
        'name': data['name'],
        'price': float(data['price']),
        'stock': int(data['stock'])
    }

    products.append(new_product)
    logger.info(f"Product created: {new_product['id']}")

    return jsonify({'success': True, 'product': new_product}), 201


@app.route('/metrics', methods=['GET'])
def metrics():
    """Prometheus metrics"""
    uptime = (datetime.now() - app_state['startup_time']).total_seconds()
    metrics_output = f"""# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{{service="product-api",method="GET"}} {random.randint(100, 1000)}

# HELP products_total Total number of products
# TYPE products_total gauge
products_total {len(products)}

# HELP app_uptime_seconds Application uptime
# TYPE app_uptime_seconds gauge
app_uptime_seconds {uptime}
"""
    return metrics_output, 200, {'Content-Type': 'text/plain'}


@app.route('/', methods=['GET'])
def index():
    return jsonify({
        'service': 'Product API',
        'version': '1.0.0',
        'endpoints': {
            'health': '/health',
            'ready': '/ready',
            'products': '/api/products',
            'metrics': '/metrics'
        }
    }), 200


if __name__ == '__main__':
    logger.info(f"Starting Product API on port {PORT}")
    app.run(host='0.0.0.0', port=PORT, debug=(ENV == 'development'))
