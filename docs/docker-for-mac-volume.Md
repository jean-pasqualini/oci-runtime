volumes:
- ./src:/app:consistent   # safest (default)
- ./src:/app:cached       # host is authoritative (good for web dev)
- ./src:/app:delegated    # container is authoritative (good for DB/data dirs)