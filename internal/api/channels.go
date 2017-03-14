package api

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
)

// ChannelListAPI exports the channel-list related functions.
type ChannelListAPI struct {
	ctx       common.Context
	validator auth.Validator
}

// NewChannelListAPI creates a new ChannelListAPI.
func NewChannelListAPI(ctx common.Context, validator auth.Validator) *ChannelListAPI {
	return &ChannelListAPI{
		ctx:       ctx,
		validator: validator,
	}
}

// Create creates the given channel-list.
func (a *ChannelListAPI) Create(ctx context.Context, req *pb.CreateChannelListRequest) (*pb.CreateChannelListResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateChannelListAccess(auth.Create)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	cl := storage.ChannelList{
		Name: req.Name,
	}

	for _, v := range req.Channels {
		cl.Channels = append(cl.Channels, int64(v))
	}

	err := storage.CreateChannelList(a.ctx.DB, &cl)
	if err != nil {
		return nil, errToRPCError(err)
	}

	return &pb.CreateChannelListResponse{Id: cl.ID}, nil
}

// Update updates the given channel-list.
func (a *ChannelListAPI) Update(ctx context.Context, req *pb.UpdateChannelListRequest) (*pb.UpdateChannelListResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateChannelListAccess(auth.Update)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	cl := storage.ChannelList{
		ID:   req.Id,
		Name: req.Name,
	}

	for _, v := range req.Channels {
		cl.Channels = append(cl.Channels, int64(v))
	}

	err := storage.UpdateChannelList(a.ctx.DB, cl)
	if err != nil {
		return nil, errToRPCError(err)
	}
	return &pb.UpdateChannelListResponse{}, nil
}

// Get returns the channel-list matching the given id.
func (a *ChannelListAPI) Get(ctx context.Context, req *pb.GetChannelListRequest) (*pb.GetChannelListResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateChannelListAccess(auth.Read)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	cl, err := storage.GetChannelList(a.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}

	var channels []uint32
	for _, v := range cl.Channels {
		channels = append(channels, uint32(v))
	}

	return &pb.GetChannelListResponse{
		Id:       cl.ID,
		Name:     cl.Name,
		Channels: channels,
	}, nil
}

// List lists the channel-lists.
func (a *ChannelListAPI) List(ctx context.Context, req *pb.ListChannelListRequest) (*pb.ListChannelListResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateChannelListAccess(auth.List)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	lists, err := storage.GetChannelLists(a.ctx.DB, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, errToRPCError(err)
	}
	count, err := storage.GetChannelListsCount(a.ctx.DB)
	if err != nil {
		return nil, errToRPCError(err)
	}

	resp := pb.ListChannelListResponse{
		TotalCount: int64(count),
	}
	for _, l := range lists {
		var cl []uint32

		for _, v := range l.Channels {
			cl = append(cl, uint32(v))
		}

		resp.Result = append(resp.Result, &pb.GetChannelListResponse{
			Id:       l.ID,
			Name:     l.Name,
			Channels: cl,
		})
	}
	return &resp, nil
}

// Delete deletes the channel-list matching the given id.
func (a *ChannelListAPI) Delete(ctx context.Context, req *pb.DeleteChannelListRequest) (*pb.DeleteChannelListResponse, error) {
	if err := a.validator.Validate(ctx,
		auth.ValidateChannelListAccess(auth.Delete)); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication failed: %s", err)
	}

	err := storage.DeleteChannelList(a.ctx.DB, req.Id)
	if err != nil {
		return nil, errToRPCError(err)
	}
	return &pb.DeleteChannelListResponse{}, nil
}
