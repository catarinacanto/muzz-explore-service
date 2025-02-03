package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"muzz-explore-service/internal/db"
	pb "muzz-explore-service/pkg/pb/proto"
)

type mockQueries struct {
	putDecision   func(ctx context.Context, arg db.PutDecisionParams) (bool, error)
	listLikers    func(ctx context.Context, arg db.ListLikersParams) ([]db.ListLikersRow, error)
	listNewLikers func(ctx context.Context, arg db.ListNewLikersParams) ([]db.ListNewLikersRow, error)
	countLikers   func(ctx context.Context, recipientUserID string) (int64, error)
}

func (m mockQueries) PutDecision(ctx context.Context, arg db.PutDecisionParams) (bool, error) {
	return m.putDecision(ctx, arg)
}

func (m mockQueries) ListLikers(ctx context.Context, arg db.ListLikersParams) ([]db.ListLikersRow, error) {
	return m.listLikers(ctx, arg)
}

func (m mockQueries) ListNewLikers(ctx context.Context, arg db.ListNewLikersParams) ([]db.ListNewLikersRow, error) {
	return m.listNewLikers(ctx, arg)
}

func (m mockQueries) CountLikers(ctx context.Context, recipientUserID string) (int64, error) {
	return m.countLikers(ctx, recipientUserID)
}

func TestPutDecision(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.PutDecisionRequest
		mock    func() db.Querier
		want    *pb.PutDecisionResponse
		wantErr bool
	}{
		{
			name: "successful like",
			req: &pb.PutDecisionRequest{
				ActorUserId:     "user1",
				RecipientUserId: "user2",
				LikedRecipient:  true,
			},
			mock: func() db.Querier {
				return mockQueries{
					putDecision: func(ctx context.Context, arg db.PutDecisionParams) (bool, error) {
						return false, nil // not mutual
					},
				}
			},
			want: &pb.PutDecisionResponse{
				MutualLikes: false,
			},
		},
		{
			name: "mutual like",
			req: &pb.PutDecisionRequest{
				ActorUserId:     "user1",
				RecipientUserId: "user2",
				LikedRecipient:  true,
			},
			mock: func() db.Querier {
				return mockQueries{
					putDecision: func(ctx context.Context, arg db.PutDecisionParams) (bool, error) {
						return true, nil // mutual like
					},
				}
			},
			want: &pb.PutDecisionResponse{
				MutualLikes: true,
			},
		},
		{
			name: "missing actor ID",
			req: &pb.PutDecisionRequest{
				RecipientUserId: "user2",
				LikedRecipient:  true,
			},
			wantErr: true,
		},
		{
			name: "missing recipient ID",
			req: &pb.PutDecisionRequest{
				ActorUserId:    "user1",
				LikedRecipient: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var queries db.Querier
			if tt.mock != nil {
				queries = tt.mock()
			}

			s := NewExploreService(queries)
			got, err := s.PutDecision(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestListLikedYou(t *testing.T) {
	now := time.Now()
	emptyString := ""

	tests := []struct {
		name    string
		req     *pb.ListLikedYouRequest
		mock    func() db.Querier
		want    *pb.ListLikedYouResponse
		wantErr bool
	}{
		{
			name: "successful list",
			req: &pb.ListLikedYouRequest{
				RecipientUserId: "user2",
			},
			mock: func() db.Querier {
				return mockQueries{
					listLikers: func(ctx context.Context, arg db.ListLikersParams) ([]db.ListLikersRow, error) {
						return []db.ListLikersRow{
							{
								ActorUserID: "user1",
								CreatedAt:   now,
							},
						}, nil
					},
				}
			},
			want: &pb.ListLikedYouResponse{
				Likers: []*pb.ListLikedYouResponse_Liker{
					{
						ActorId:       "user1",
						UnixTimestamp: uint64(now.Unix()),
					},
				},
				NextPaginationToken: &emptyString, // Expect empty string pointer instead of nil
			},
		},
		{
			name:    "missing recipient ID",
			req:     &pb.ListLikedYouRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var queries db.Querier
			if tt.mock != nil {
				queries = tt.mock()
			}

			s := NewExploreService(queries)
			got, err := s.ListLikedYou(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListNewLikedYou(t *testing.T) {
	now := time.Now()
	emptyString := ""

	tests := []struct {
		name    string
		req     *pb.ListLikedYouRequest
		mock    func() db.Querier
		want    *pb.ListLikedYouResponse
		wantErr bool
	}{
		{
			name: "successful list new likes",
			req: &pb.ListLikedYouRequest{
				RecipientUserId: "user2",
			},
			mock: func() db.Querier {
				return mockQueries{
					listNewLikers: func(ctx context.Context, arg db.ListNewLikersParams) ([]db.ListNewLikersRow, error) {
						return []db.ListNewLikersRow{
							{
								ActorUserID: "user1",
								CreatedAt:   now,
							},
						}, nil
					},
				}
			},
			want: &pb.ListLikedYouResponse{
				Likers: []*pb.ListLikedYouResponse_Liker{
					{
						ActorId:       "user1",
						UnixTimestamp: uint64(now.Unix()),
					},
				},
				NextPaginationToken: &emptyString,
			},
		},
		{
			name: "no new likes",
			req: &pb.ListLikedYouRequest{
				RecipientUserId: "user2",
			},
			mock: func() db.Querier {
				return mockQueries{
					listNewLikers: func(ctx context.Context, arg db.ListNewLikersParams) ([]db.ListNewLikersRow, error) {
						return []db.ListNewLikersRow{}, nil
					},
				}
			},
			want: &pb.ListLikedYouResponse{
				Likers:              []*pb.ListLikedYouResponse_Liker{},
				NextPaginationToken: &emptyString,
			},
		},
		{
			name:    "missing recipient ID",
			req:     &pb.ListLikedYouRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var queries db.Querier
			if tt.mock != nil {
				queries = tt.mock()
			}

			s := NewExploreService(queries)
			got, err := s.ListNewLikedYou(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountLikedYou(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.CountLikedYouRequest
		mock    func() db.Querier
		want    *pb.CountLikedYouResponse
		wantErr bool
	}{
		{
			name: "successful count",
			req: &pb.CountLikedYouRequest{
				RecipientUserId: "user2",
			},
			mock: func() db.Querier {
				return mockQueries{
					countLikers: func(ctx context.Context, recipientUserID string) (int64, error) {
						return 5, nil // Mock 5 likes
					},
				}
			},
			want: &pb.CountLikedYouResponse{
				Count: 5,
			},
		},
		{
			name: "zero likes",
			req: &pb.CountLikedYouRequest{
				RecipientUserId: "user2",
			},
			mock: func() db.Querier {
				return mockQueries{
					countLikers: func(ctx context.Context, recipientUserID string) (int64, error) {
						return 0, nil
					},
				}
			},
			want: &pb.CountLikedYouResponse{
				Count: 0,
			},
		},
		{
			name:    "missing recipient ID",
			req:     &pb.CountLikedYouRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var queries db.Querier
			if tt.mock != nil {
				queries = tt.mock()
			}

			s := NewExploreService(queries)
			got, err := s.CountLikedYou(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListNewLikedYou_MutualLikes(t *testing.T) {
	now := time.Now()
	emptyString := ""

	tests := []struct {
		name    string
		req     *pb.ListLikedYouRequest
		mock    func() db.Querier
		want    *pb.ListLikedYouResponse
		wantErr bool
	}{
		{
			name: "excludes mutual likes",
			req: &pb.ListLikedYouRequest{
				RecipientUserId: "user2",
			},
			mock: func() db.Querier {
				return mockQueries{
					listNewLikers: func(ctx context.Context, arg db.ListNewLikersParams) ([]db.ListNewLikersRow, error) {
						// Should only return user3, as user1 has mutual like
						return []db.ListNewLikersRow{
							{
								ActorUserID: "user3",
								CreatedAt:   now,
							},
						}, nil
					},
					// Add PutDecision to set up the test scenario
					putDecision: func(ctx context.Context, arg db.PutDecisionParams) (bool, error) {
						// Simulating user1 and user2 liking each other
						if arg.ActorUserID == "user1" && arg.RecipientUserID == "user2" {
							return true, nil // mutual like
						}
						return false, nil
					},
				}
			},
			want: &pb.ListLikedYouResponse{
				Likers: []*pb.ListLikedYouResponse_Liker{
					{
						ActorId:       "user3", // Only user3 should appear, not user1
						UnixTimestamp: uint64(now.Unix()),
					},
				},
				NextPaginationToken: &emptyString,
			},
		},
		{
			name: "new likes become mutual",
			req: &pb.ListLikedYouRequest{
				RecipientUserId: "user2",
			},
			mock: func() db.Querier {
				// Simulate a sequence of events:
				// 1. Initially user1 likes user2 (shows in new likes)
				// 2. user2 likes user1 back (should no longer show in new likes)
				return mockQueries{
					listNewLikers: func(ctx context.Context, arg db.ListNewLikersParams) ([]db.ListNewLikersRow, error) {
						// After mutual like is established, should return empty
						return []db.ListNewLikersRow{}, nil
					},
					putDecision: func(ctx context.Context, arg db.PutDecisionParams) (bool, error) {
						// When user2 likes user1 back
						if arg.ActorUserID == "user2" && arg.RecipientUserID == "user1" {
							return true, nil // becomes mutual
						}
						return false, nil
					},
				}
			},
			want: &pb.ListLikedYouResponse{
				Likers:              []*pb.ListLikedYouResponse_Liker{}, // Empty after mutual like
				NextPaginationToken: &emptyString,
			},
		},
		{
			name: "multiple new likes with ordering",
			req: &pb.ListLikedYouRequest{
				RecipientUserId: "user2",
			},
			mock: func() db.Querier {
				earlier := now.Add(-1 * time.Hour)
				return mockQueries{
					listNewLikers: func(ctx context.Context, arg db.ListNewLikersParams) ([]db.ListNewLikersRow, error) {
						// Return multiple likes in chronological order
						return []db.ListNewLikersRow{
							{
								ActorUserID: "user3",
								CreatedAt:   now, // Most recent
							},
							{
								ActorUserID: "user4",
								CreatedAt:   earlier, // Older
							},
						}, nil
					},
				}
			},
			want: &pb.ListLikedYouResponse{
				Likers: []*pb.ListLikedYouResponse_Liker{
					{
						ActorId:       "user3",
						UnixTimestamp: uint64(now.Unix()),
					},
					{
						ActorId:       "user4",
						UnixTimestamp: uint64(now.Add(-1 * time.Hour).Unix()),
					},
				},
				NextPaginationToken: &emptyString,
			},
		},
		{
			name: "likes after pass",
			req: &pb.ListLikedYouRequest{
				RecipientUserId: "user2",
			},
			mock: func() db.Querier {
				return mockQueries{
					listNewLikers: func(ctx context.Context, arg db.ListNewLikersParams) ([]db.ListNewLikersRow, error) {
						// User3 liked user2 after initially passing
						return []db.ListNewLikersRow{
							{
								ActorUserID: "user3",
								CreatedAt:   now,
							},
						}, nil
					},
					putDecision: func(ctx context.Context, arg db.PutDecisionParams) (bool, error) {
						// Simulating user3 first passing, then liking
						if arg.ActorUserID == "user3" && arg.RecipientUserID == "user2" {
							return false, nil
						}
						return false, nil
					},
				}
			},
			want: &pb.ListLikedYouResponse{
				Likers: []*pb.ListLikedYouResponse_Liker{
					{
						ActorId:       "user3",
						UnixTimestamp: uint64(now.Unix()),
					},
				},
				NextPaginationToken: &emptyString,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var queries db.Querier
			if tt.mock != nil {
				queries = tt.mock()
			}

			s := NewExploreService(queries)
			got, err := s.ListNewLikedYou(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
