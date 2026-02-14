package greeter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/greeter-gripmock-embedded/helloworld"
	sdk "github.com/bavix/gripmock/v3/pkg/sdk"
)

// setupGreeterMock creates an embedded GripMock with the greeter service and returns mock + client.
func setupGreeterMock(t *testing.T) (sdk.Mock, helloworld.GreeterClient) {
	t.Helper()
	ctx := t.Context()
	mock, err := sdk.Run(ctx, sdk.WithFileDescriptor(helloworld.File_service_proto))
	require.NoError(t, err)
	t.Cleanup(func() { _ = mock.Close() })

	return mock, helloworld.NewGreeterClient(mock.Conn())
}

func TestGreeter_SayHello(t *testing.T) {
	t.Parallel()

	// Arrange
	mock, client := setupGreeterMock(t)
	mock.Stub("helloworld.Greeter", "SayHello").
		Unary("name", "Bob", "message", "Hello Bob").
		Commit()

	// Act
	reply, err := client.SayHello(t.Context(), &helloworld.HelloRequest{Name: "Bob"})

	// Assert
	require.NoError(t, err)
	require.Equal(t, "Hello Bob", reply.GetMessage())
}

func TestGreeter_SayHello_DynamicTemplate(t *testing.T) {
	t.Parallel()

	// Arrange
	mock, client := setupGreeterMock(t)
	mock.Stub("helloworld.Greeter", "SayHello").
		When(sdk.Matches("name", ".+")).
		Return("message", "Hi {{.Request.name}}").
		Commit()

	// Act
	reply, err := client.SayHello(t.Context(), &helloworld.HelloRequest{Name: "Alex"})

	// Assert
	require.NoError(t, err)
	require.Equal(t, "Hi Alex", reply.GetMessage())
}
