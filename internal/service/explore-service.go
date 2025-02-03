package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"muzz-explore-service/internal/db"
	pb "muzz-explore-service/pkg/pb/proto"
	"time"
)

const pageSize = 50

type ExploreService struct {
	pb.UnimplementedExploreServiceServer
	queries db.Querier
}

func NewExploreService(queries db.Querier) *ExploreService {
	return &ExploreService{
		queries: queries,
	}
}

// PutDecision records a user's decision to like or pass another user
// Returns whether this creates a mutual like between the users
func (s *ExploreService) PutDecision(ctx context.Context, req *pb.PutDecisionRequest) (*pb.PutDecisionResponse, error) {
	if req.ActorUserId == "" || req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "both actor_user_id and recipient_user_id are required")
	}

	// Prevent self-liking
	if req.ActorUserId == req.RecipientUserId {
		return nil, status.Error(codes.InvalidArgument, "users can't like themselves")
	}

	mutualLikes, err := s.queries.PutDecision(ctx, db.PutDecisionParams{
		ActorUserID:     req.ActorUserId,
		RecipientUserID: req.RecipientUserId,
		Liked:           req.LikedRecipient,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to record decision")
	}

	return &pb.PutDecisionResponse{
		MutualLikes: mutualLikes,
	}, nil
}

// ListLikedYou returns a list of users who have liked the recipient
// Supports pagination using cursor-based pagination for efficiency
func (s *ExploreService) ListLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	if req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient_user_id is required")
	}

	cursorTime, err := s.decodePaginationToken(req.PaginationToken)
	if err != nil {
		return nil, err
	}

	params := db.ListLikersParams{
		RecipientUserID: req.RecipientUserId,
		CreatedAtCursor: cursorTime,
		PageLimit:       int32(pageSize + 1), // Fetch one extra item to check for next page
	}

	decisions, err := s.queries.ListLikers(ctx, params)
	if err != nil {
		log.Printf("Error fetching decisions: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch likers")
	}
	log.Printf("Found %d decisions", len(decisions))

	// Generate next page token if we have more results
	var nextToken string
	if len(decisions) > pageSize {
		nextToken, err = s.generateNextToken(decisions[pageSize-1].CreatedAt)
		if err != nil {
			return nil, err
		}
		decisions = decisions[:pageSize] // Truncate decisions to the page size
	}

	// Convert database results to protobuf response format
	likers := make([]*pb.ListLikedYouResponse_Liker, len(decisions))
	for i, d := range decisions {
		likers[i] = &pb.ListLikedYouResponse_Liker{
			ActorId:       d.ActorUserID,
			UnixTimestamp: uint64(d.CreatedAt.Unix()),
		}
	}

	return &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextToken,
	}, nil
}

// ListNewLikedYou returns a list of users who have liked the recipient but haven't been liked back
// Similar to ListLikedYou but excludes mutual likes
func (s *ExploreService) ListNewLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	if req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient_user_id is required")
	}

	cursorTime, err := s.decodePaginationToken(req.PaginationToken)
	if err != nil {
		return nil, err
	}

	params := db.ListNewLikersParams{
		RecipientUserID: req.RecipientUserId,
		CreatedAtCursor: cursorTime,
		PageLimit:       int32(pageSize + 1), // Fetch one extra item to check for next page
	}
	log.Printf("Query params: %+v", params)

	decisions, err := s.queries.ListNewLikers(ctx, params)
	if err != nil {
		log.Printf("Error fetching decisions: %v", err)
		return nil, status.Error(codes.Internal, "failed to fetch new likers")
	}
	log.Printf("Found %d decisions", len(decisions))

	// Generate next page token if we have more results
	var nextToken string
	if len(decisions) > pageSize {
		nextToken, err = s.generateNextToken(decisions[pageSize-1].CreatedAt)
		if err != nil {
			return nil, err
		}
		decisions = decisions[:pageSize] // Truncate decisions to the page size
	}

	// Convert database results to protobuf response format
	likers := make([]*pb.ListLikedYouResponse_Liker, len(decisions))
	for i, d := range decisions {
		likers[i] = &pb.ListLikedYouResponse_Liker{
			ActorId:       d.ActorUserID,
			UnixTimestamp: uint64(d.CreatedAt.Unix()),
		}
	}

	return &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextToken,
	}, nil
}

// CountLikedYou returns the total number of users who have liked the recipient
func (s *ExploreService) CountLikedYou(ctx context.Context, req *pb.CountLikedYouRequest) (*pb.CountLikedYouResponse, error) {
	if req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient_user_id is required")
	}

	count, err := s.queries.CountLikers(ctx, req.RecipientUserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to count likers")
	}

	return &pb.CountLikedYouResponse{
		Count: uint64(count),
	}, nil
}

// decodePaginationToken decodes a base64-encoded pagination token into a timestamp.
// It is used to determine the starting point for cursor-based pagination.
// Returns the decoded timestamp.
func (s *ExploreService) decodePaginationToken(token *string) (time.Time, error) {
	if token == nil || *token == "" {
		return time.Time{}, nil // Return zero time if no token is provided
	}

	data, err := base64.StdEncoding.DecodeString(*token)
	if err != nil {
		return time.Time{}, status.Error(codes.InvalidArgument, "invalid pagination token")
	}

	var timestamp int64
	if err := json.Unmarshal(data, &timestamp); err != nil {
		return time.Time{}, status.Error(codes.InvalidArgument, "invalid pagination token")
	}

	return time.Unix(timestamp, 0), nil
}

// generateNextToken generates a base64-encoded pagination token from a timestamp.
// It is used to create a token for the next page of results.
func (s *ExploreService) generateNextToken(lastTime time.Time) (string, error) {
	data, err := json.Marshal(lastTime.Unix())
	if err != nil {
		return "", status.Error(codes.Internal, "failed to generate pagination token")
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
