package internal

import (
	"context"
	"log"
	"time"

	pb "github.com/assidiqi598/erp/services/auth/proto"
	"github.com/assidiqi598/erp/shared/auth"
	"github.com/assidiqi598/erp/shared/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Implement Login method
func (s *AuthServer) LoginWithEmailAndPass(ctx context.Context, req *pb.LoginWithEmailAndPassRequest) (*pb.LoginResponse, error) {
	repo := repositories.NewUserRepository()

	// Fetch user from MongoDB
	user, err := repo.FindUser(bson.M{"email": req.Email})
	if err != nil {
		log.Printf("User not found: %v", err)
		return nil, status.Errorf(codes.NotFound, "User not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))

	if err != nil {
		log.Printf("Invalid password for user: %v", user.Email)
		return nil, status.Errorf(codes.Unauthenticated, "Invalid credentials")
	}

	// Generate tokens
	accessToken, err := auth.GenerateJWT(user.ID.Hex(), user.Email, user.PhoneNumber, time.Minute*60)
	if err != nil {
		log.Printf("Error generating access token: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to generate access token")
	}

	refreshToken, err := auth.GenerateJWT(user.ID.Hex(), user.Email, user.PhoneNumber, time.Hour*24)
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to generate refresh token")
	}

	err = repo.UpdateUser(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{
			"$set": bson.M{
				"last_login": time.Now(),
			},
		})

	if err != nil {
		log.Printf("Error updating user: %v", err)
	}

	// Return a successful response
	return &pb.LoginResponse{Token: accessToken, Message: "Login successful", RefreshToken: refreshToken}, nil
}
