#!/bin/bash

# Function to create likes
create_likes() {
    for i in {1..5}; do
        echo "Creating like from tester$i to user2..."
        grpcurl -plaintext -d "{
            \"actor_user_id\": \"tester$i\",
            \"recipient_user_id\": \"user2\",
            \"liked_recipient\": true
        }" localhost:8080 explore.ExploreService/PutDecision
    done
}

# Function to get likes
get_likes() {
    echo -e "\nGetting likes for user2..."
    grpcurl -plaintext -d '{
        "recipient_user_id": "user2"
    }' localhost:8080 explore.ExploreService/ListLikedYou
}

# Main test flow
echo "Starting pagination test..."
create_likes
get_likes

echo -e "\nTo test with pagination token, copy the nextPaginationToken from above and run:"
echo 'grpcurl -plaintext -d '"'"'{"recipient_user_id": "user2", "pagination_token": "PASTE_TOKEN_HERE"}'"'"' localhost:8080 explore.ExploreService/ListLikedYou'