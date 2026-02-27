package cognito

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// Client is the interface used by the users service for Cognito operations.
// Implementations can be the real AWS client or a test fake.
type Client interface {
	SignUp(ctx context.Context, email, password string) (sub string, err error)
	InitiateAuth(ctx context.Context, email, password string) (sub string, err error)
	AdminDeleteUser(ctx context.Context, email string) error
}

// AWSClient implements Client using the AWS Cognito Identity Provider SDK.
type AWSClient struct {
	svc        *cognitoidentityprovider.Client
	userPoolID string
	clientID   string
}

// NewAWSClient creates an AWS-backed Cognito client.
// region is typically the same as AWS_REGION.
func NewAWSClient(ctx context.Context, region, userPoolID, clientID string, cfgFns ...func(*config.LoadOptions) error) (*AWSClient, error) {
	if userPoolID == "" || clientID == "" {
		return nil, errors.New("cognito: userPoolID and clientID are required")
	}
	loadOpts := []func(*config.LoadOptions) error{config.WithRegion(region)}
	loadOpts = append(loadOpts, cfgFns...)
	awsCfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, err
	}
	svc := cognitoidentityprovider.NewFromConfig(awsCfg)
	return &AWSClient{
		svc:        svc,
		userPoolID: userPoolID,
		clientID:   clientID,
	}, nil
}

// SignUp creates a user in the Cognito User Pool with the given email and password and returns the Cognito sub.
// This uses AdminCreateUser + AdminSetUserPassword so the backend fully controls confirmation.
func (c *AWSClient) SignUp(ctx context.Context, email, password string) (string, error) {
	username := email

	createOut, err := c.svc.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(username),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
		},
		// Suppress welcome email; we control UX.
		MessageAction: types.MessageActionTypeSuppress,
	})
	if err != nil {
		return "", err
	}

	_, err = c.svc.AdminSetUserPassword(ctx, &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(username),
		Password:   aws.String(password),
		Permanent:  true,
	})
	if err != nil {
		return "", err
	}

	var sub string
	if createOut.User != nil {
		for _, attr := range createOut.User.Attributes {
			if aws.ToString(attr.Name) == "sub" {
				sub = aws.ToString(attr.Value)
				break
			}
		}
	}
	if sub == "" {
		// Fallback: fetch user to get sub.
		getOut, err := c.svc.AdminGetUser(ctx, &cognitoidentityprovider.AdminGetUserInput{
			UserPoolId: aws.String(c.userPoolID),
			Username:   aws.String(username),
		})
		if err != nil {
			return "", err
		}
		for _, attr := range getOut.UserAttributes {
			if aws.ToString(attr.Name) == "sub" {
				sub = aws.ToString(attr.Value)
				break
			}
		}
	}
	if sub == "" {
		return "", errors.New("cognito: sub not found on created user")
	}
	return sub, nil
}

// InitiateAuth verifies the user's credentials and returns their Cognito sub.
func (c *AWSClient) InitiateAuth(ctx context.Context, email, password string) (string, error) {
	username := email

	_, err := c.svc.AdminInitiateAuth(ctx, &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow: types.AuthFlowTypeAdminUserPasswordAuth,
		ClientId: aws.String(c.clientID),
		UserPoolId: aws.String(c.userPoolID),
		AuthParameters: map[string]string{
			"USERNAME": username,
			"PASSWORD": password,
		},
	})
	if err != nil {
		return "", err
	}

	getOut, err := c.svc.AdminGetUser(ctx, &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(username),
	})
	if err != nil {
		return "", err
	}
	for _, attr := range getOut.UserAttributes {
		if aws.ToString(attr.Name) == "sub" {
			return aws.ToString(attr.Value), nil
		}
	}
	return "", errors.New("cognito: sub not found for user")
}

// AdminDeleteUser deletes the user from the Cognito User Pool by email (username).
func (c *AWSClient) AdminDeleteUser(ctx context.Context, email string) error {
	if email == "" {
		return nil
	}
	_, err := c.svc.AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(email),
	})
	return err
}

