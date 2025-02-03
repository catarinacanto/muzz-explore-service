

# Muzz Explore Service

A gRPC service that manages user interactions (likes/passes) on Muzz. This service handles user decisions and provides endpoints to view and manage likes between users.

## Features

- Record user decisions (likes/passes)
- List users who liked a specific user
- List new (non-mutual) likes for a user
- Count total likes for a user
- Supports pagination for large datasets
- Handles mutual likes detection

## Assumptions

- User IDs are unique and valid
- Users can't like themselves
- A decision (like/pass) can be changed at any time
- There's no time limit on when users can like each other
- Mutual likes are determined by both users liking each other, regardless of timing
- Pagination tokens are immutable and secure (using base64 encoding)

## Technical Stack

- Go 1.23
- PostgreSQL 16
- gRPC
- SQLc for type-safe database queries
- Docker for containerisation

## Getting Started

### Prerequisites

- Go 1.23 or later
- Docker and Docker Compose

### Setup

#### Clone the repository:

```bash
git clone git@github.com:catarinacanto/muzz-explore-service.git && cd muzz-explore-service
```  

#### Install dependencies:

```bash
go mod tidy
```

#### Start the service:
```bash
`docker compose up --build`
```

### Running Tests
```bash
`go test ./... -v`
```

## API Endpoints

-   `PutDecision`: Record a user's decision to like or pass another user
    - Returns whether the like is mutual
-   `ListLikedYou`: List all users who liked the recipient
    - Supports pagination
    - Returns timestamp of like
-   `ListNewLikedYou`: List users who liked the recipient (excluding mutual likes)
    - Supports pagination
    - Returns timestamp of like
-   `CountLikedYou`: Count the number of users who liked the recipient

## Design Decisions

- Uses cursor-based pagination for efficient handling of large datasets
- PostgreSQL for reliable ACID transactions and complex queries
- SQLc for type-safe database operations
- Containerized for consistent development and deployment

## Testing

The service can be tested using grpcurl:
#### gRPC Debugging
- Install grpcurl for gRPC service testing:
  ```bash
  go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

### 1. Put Decision (like)

```bash
    grpcurl -plaintext -d '{  
    "actor_user_id": "user1",  
    "recipient_user_id": "user2",  
    "liked_recipient": true  
    }' localhost:8080 explore.ExploreService/PutDecision  
```


### 2. Put Decision (pass)
```bash
    grpcurl -plaintext -d '{  
    "actor_user_id": "user1",  
    "recipient_user_id": "user2",  
    "liked_recipient": false  
    }' localhost:8080 explore.ExploreService/PutDecision  
```

### 3. List all likes for user2
```bash
    grpcurl -plaintext -d '{  
    "recipient_user_id": "user2"  
    }' localhost:8080 explore.ExploreService/ListLikedYou  
```

### 4. List new (non-mutual) likes for user2
```bash
    grpcurl -plaintext -d '{  
    "recipient_user_id": "user2"  
    }' localhost:8080 explore.ExploreService/ListNewLikedYou  
```

### 5. Count likes for user2
```bash
    grpcurl -plaintext -d '{  
    "recipient_user_id": "user2"  
    }' localhost:8080 explore.ExploreService/CountLikedYou  
```

### 6. List likes with pagination
```bash
    grpcurl -plaintext -d '{  
    "recipient_user_id": "user2",  
    "pagination_token": "PASTE_TOKEN_HERE"  
    }' localhost:8080 explore.ExploreService/ListLikedYou  
```

### 7. List new likes with pagination
```bash
    grpcurl -plaintext -d '{  
    "recipient_user_id": "user2",  
    "pagination_token": "PASTE_TOKEN_HERE"  
    }' localhost:8080 explore.ExploreService/ListNewLikedYou  
```

## Scaling Considerations

- Cursor-based pagination for efficient handling of large datasets
- Index optimization for common queries
