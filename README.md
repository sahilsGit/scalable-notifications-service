# Fully-Scalable De-coupled Notifications Service

A Conceptual highly scalable, loosely coupled distributed notification system designed for high throughput and reliability.

## System Overview

This project implements a minimal concept as to how notifications can be handled at scale. Built using a microservices architecture with event driven communication patterns, it handles reliable delivery across multiple channels with minimal coupling between components. 

For simplicity all the services are managed by a single `docker-compose` file, but in real-world scenario it can orchestrated using `kubernetes` or something.

## Architecture Diagram

```mermaid
graph TB
    Client["Client Applications"]:::external
    
    subgraph "Entry Layer"
        EnqueueSvc["Enqueue Service (Go, REST API)"]:::service
    end
    
    subgraph "Kafka Backbone"
        RawTopic["Raw Notifications Topic"]:::kafka
        HighTopic["High Priority Topic"]:::kafka
        MediumTopic["Medium Priority Topic"]:::kafka
        LowTopic["Low Priority Topic"]:::kafka
        DeliveryTopic["Delivery Topics (email, push,
         etc)"]:::kafka
    end
    
    subgraph "Processing Layer"
        Validator["Validator & Prioritizer"]:::service
        Preferences-RateLimiter["Preferences & Rate Limiter"]:::service
        Tracker["Notification Tracker"]:::service
    end
    
    subgraph "Delivery Layer"
        EmailHandler["Email Handler"]:::service
        PushHandler["Push Notification Handler"]:::service
        OtherHandlers["Other Channel Handlers"]:::service
    end
    
    subgraph "Data Stores"
        Redis["Redis (Rate Limiting)"]:::database
        MySQL["MySQL (User Preferences)"]:::database
        Cassandra["Cassandra (Notification 
        History)"]:::database
    end
    
    Client -->|Generate event| EnqueueSvc
    EnqueueSvc -->|Publish raw event| RawTopic
    RawTopic -->|Consume| Validator

    
    Validator -->|High priority| HighTopic
    Validator -->|Medium priority| MediumTopic
    Validator -->|Low priority| LowTopic
    
    HighTopic & MediumTopic & LowTopic -->|Consume by priority| Preferences-RateLimiter
    Preferences-RateLimiter <-->|Query/Update| Redis
    Preferences-RateLimiter -->|Publish to delivery| DeliveryTopic
    Preferences-RateLimiter <-->|Query| MySQL
    
    DeliveryTopic -->|Route to handlers| EmailHandler & PushHandler & OtherHandlers
    
    DeliveryTopic -.->|Record history| Tracker
    Tracker -.->|Store| Cassandra
    
    EmailHandler -->|Send| Client
    PushHandler -->|Notify| Client
    OtherHandlers -->|Deliver| Client
```

## Key Features

- ✅ **Microservices Architecture**: Loosely coupled services; Kafka being the heart of the system

- ✅ **Priority-Based Processing**: Different processing lanes for different notification priorities
- ✅ **Rate Limiting**: Redis-backed rate limiting skeleton to prevent notification fatigue & possible DDoS attacks
- ✅ **Horizontal Scalability**: Each component can be independently scaled
- ✅ **Fault Tolerance**: Resilient design with enough redundancy to handle broker failures
- ✅ **Event Tracking**: Cassandra-backed notification history Skeleton for analytics and auditing

## Architecture Components

- __**Enqueue Service**__: Entry point for all notification requests. Validates and publishes events to Kafka.
- __**Notification Validator & Prioritizer Service**__: Consumes, validates, assigns priorities, and dispatches to appropriate topic.
- __**Rate Limiter Service**__: Controls notification flow and applies rate limiting.

- __**Notification Tracker (Future Plan)**__: Records notification history for analytics and auditing (SKELETON)
- **Data Stores**: 
  - **Redis**: For rate limiting
  - **MySQL**: For user preferences
  - **Cassandra**: For notification history (SKELETON)

## Data Flow

```mermaid
sequenceDiagram
    participant Client as Client App
    participant NotifSvc as Notification Service
    participant Kafka1 as Kafka (Raw)
    participant Validator as Validator & Prioritizer
    participant PrefSvc as Preferences Service
    participant Kafka2 as Kafka (Priority Topics)
    participant RateLimit as Rate Limiter
    participant Kafka3 as Kafka (Delivery)
    participant Handlers as Delivery Handlers
    
    Client->>NotifSvc: Generate event (like, comment, etc)
    NotifSvc->>Kafka1: Publish raw event
    Kafka1->>Validator: Consume event
    Validator->>PrefSvc: Check user preferences
    PrefSvc-->>Validator: Return preferences
    
    Validator->>Validator: Validate & assign priority
    Validator->>Kafka2: Publish to appropriate priority topic
    
    Kafka2->>RateLimit: Consume by priority
    RateLimit->>RateLimit: Apply rate limits
    RateLimit->>Kafka3: Publish to delivery topics
    
    Kafka3->>Handlers: Route to appropriate handler
    
    alt Email Notification
        Handlers->>Client: Send email
    end
```

## Scaling Strategy

The system is designed to scale horizontally at each layer:

```mermaid
flowchart LR
    subgraph "Producer Layer"
        P1["Client App 1"]
        P2["Client App 2"]
        P3["Client App n"]
    end
    
    subgraph "Notification Service Layer"
        N1["Notification Service 1"]
        N2["Notification Service 2"]
        N3["Notification Service n"]
    end
    
    subgraph "Kafka Raw Events"
        K1["Broker 1"]
        K2["Broker 2"]
        K3["Broker n"]
    end
    
    subgraph "Validation Layer"
        V1["Validator 1"]
        V2["Validator 2"]
        V3["Validator n"]
    end
    
    subgraph "Rate Limiting Layer"
        R1["Rate Limiter 1"]
        R2["Rate Limiter 2"]
        R3["Rate Limiter n"]
    end
    
    subgraph "Handler Layer"
        H1["Handler Group 1"]
        H2["Handler Group 2"]
        H3["Handler Group n"]
    end
    
    P1 & P2 & P3 --> N1 & N2 & N3
    N1 & N2 & N3 --> K1 & K2 & K3
    K1 & K2 & K3 --> V1 & V2 & V3
    V1 & V2 & V3 --> R1 & R2 & R3
    R1 & R2 & R3 --> H1 & H2 & H3
```

Each component can be independently scaled based on load patterns, with Kafka partitioning ensuring parallel processing.

## Design Decisions and Trade-offs

### Why Kafka?
Kafka for high-throughput, persistent event streaming needs. Its partitioning, consumer-groups and leader-follower paradigm opens up path to easy horizontal scaling and fault tolerance.

### Why Multiple Services?
Breaking the system into microservices allows for:
- Independent scaling based on load patterns
- Isolated failure domains
- Easy management and debugging

### Rate Limiting Strategy
Rate limiting can be implemented at multiple levels:
- Global system-wide limits
- Per-user limits
- Per-channel limits
- Priority-based limits

## Example Usage

- Spin up the services using `docker compose up` in /`infrastructure` directory. 
- Once the services are up, use example curl below.

`
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-001",
    "event_type": "security_alert",
    "content": "Suspicious login attempt detected from a new device",
    "metadata": {
      "device": "Unknown iPhone",
      "location": "Somewhere in the world",
      "ip_address": "203.0.113.42",
      "timestamp": '$(date +%s)'
    }
  }'
`
- Then use `docker exec -it kafka-1 kafka-console-consumer --bootstrap-server localhost:9092 --topic notifications.raw --from-beginning` to see the messages in the topic.
- You can also use `docker-compose -f notifications-service/infrastructure/docker-compose.yml logs -f [<service-name>]` to see the logs .

- If you see issues such as service exited during subsequent startup, make sure to remove old volumes using command such as `docker volume rm [<volume_name>]`.