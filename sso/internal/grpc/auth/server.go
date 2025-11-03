package auth

import (
	"context"
	"errors"

	"github.com/asaskevich/govalidator"
	"github.com/lostmyescape/link-shortener/sso/internal/services/auth"
	ssov1 "github.com/lostmyescape/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	Register(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {

	if err := validateLogin(req); err != nil {
		return nil, err
	}
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {

	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func validEmail(email string) bool {
	return govalidator.IsEmail(email)
}

func validPass(password string) bool {
	if password == "" {
		return false
	}
	if len(password) < 6 {
		return false
	}
	return true
}

func validateLogin(req *ssov1.LoginRequest) error {
	if !validEmail(req.GetEmail()) {
		if req.GetEmail() == "" {
			return status.Error(codes.InvalidArgument, "email is required")
		}
		return status.Error(codes.InvalidArgument, "invalid email format")
	}

	if !validPass(req.GetPassword()) {
		if req.GetPassword() == "" {
			return status.Error(codes.InvalidArgument, "password is required")
		}
		return status.Error(codes.InvalidArgument, "password must be at least 6 characters long")

	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if !validEmail(req.GetEmail()) {
		if req.GetEmail() == "" {
			return status.Error(codes.InvalidArgument, "email is required")
		}
		return status.Error(codes.InvalidArgument, "invalid email format")
	}

	if !validPass(req.GetPassword()) {
		if req.GetPassword() == "" {
			return status.Error(codes.InvalidArgument, "password is required")
		}
		return status.Error(codes.InvalidArgument, "password must be at least 6 characters long")

	}

	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	return nil
}
