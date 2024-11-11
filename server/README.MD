# Backend

## Architecture Overview
1. Language & Framework
   - Go (Golang)
   - Chi Router for HTTP routing
   - WebSocket for real-time communication
   - PostgreSQL for data persistence
   - Redis for real-time features


2. Core Components
   a. Authentication Service (`internal/auth/`)
      - JWT-based authentication
      - User registration and login
      - Password hashing with bcrypt
      - Middleware for route protection

   b. Chat Service (`internal/chat/`)
      - WebSocket connections
      - Message persistence
      - Real-time message delivery
      - Redis pub/sub for scaling

   c. User Service (`internal/user/`)
      - User management
      - User search
      - Profile management


## Key Concepts & Design Patterns

### Service-Based Organization

* **Auth Service:** Manages authentication and authorization
* **User Service:** Handles user management and search
* **Chat Service:** Manages messaging and real-time communication

Each service is self-contained with its own business logic, handlers, and domain types, promoting modularity and maintainability.

### Dependency Injection

* Improve testability by allowing mock dependencies
* Make dependencies explicit and visible
* Control the lifecycle of dependencies
* Enable better separation of concerns
* Make the codebase more maintainable and flexible

### Middleware Pattern

* Authenticate and authorize requests
* Log request/response information
* Handle cross-origin resource sharing (CORS)
* Recover from panics
* Process requests before they reach handlers

## Development Tools

### Swagger for API Documentation

* Provide interactive API documentation
* Auto-generate API client code
* Ensure API documentation stays in sync with code
* Make API testing and exploration easier
* Generate TypeScript types for frontend consumption

### Database Migrations
Database migrations are handled using golang-migrate to:

* Version control database schema changes
* Enable safe schema updates
* Provide rollback capabilities
* Maintain database consistency
* Track database change history
* Enable team collaboration on database changes

### Docker for Development

* Create a consistent development environment
* Isolate development dependencies
* Run services like PostgreSQL and Redis
* Make project setup easier for new developers
* Enable easy service management
* Provide production-like environment locally
