package main

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bavix/greeter-gripmock-embedded/helloworld"
	"github.com/bavix/greeter-gripmock-embedded/timed"
	sdk "github.com/bavix/gripmock/v3/pkg/sdk"
)

func runGreeterMock(t *testing.T) (sdk.Mock, helloworld.GreeterClient) {
	t.Helper()
	mock, err := sdk.Run(t, sdk.WithFileDescriptor(helloworld.File_greeter_proto))
	require.NoError(t, err)

	return mock, helloworld.NewGreeterClient(mock.Conn())
}

func runTimedServer(t *testing.T, greeterClient helloworld.GreeterClient) timed.TimedGreeterClient {
	t.Helper()
	lis, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	timed.RegisterTimedGreeterServer(srv, NewTimedGreeterServer(greeterClient))

	go func() { _ = srv.Serve(lis) }()

	t.Cleanup(srv.GracefulStop)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return timed.NewTimedGreeterClient(conn)
}

func setupTimedServer(t *testing.T) (timed.TimedGreeterClient, sdk.Mock) {
	t.Helper()
	mock, greeterClient := runGreeterMock(t)

	return runTimedServer(t, greeterClient), mock
}

func TestTimedGreeter_SayHello(t *testing.T) {
	t.Parallel()

	// Arrange
	client, mock := setupTimedServer(t)
	delayMs := 20
	mock.Stub("helloworld.Greeter", "SayHello").
		Unary("name", "Bob", "message", "Hello Bob").
		Delay(time.Duration(delayMs) * time.Millisecond).
		Commit()

	// Act
	reply, err := client.SayHello(t.Context(), &timed.HelloRequest{Name: "Bob"})

	// Assert
	require.NoError(t, err)
	require.Equal(t, "Hello Bob", reply.GetMessage())
	require.GreaterOrEqual(t, reply.GetDurationMs(), int64(delayMs))
}

func TestTimedGreeter_SayHello_DynamicTemplate(t *testing.T) {
	t.Parallel()

	// Arrange
	client, mock := setupTimedServer(t)
	delayMs := 30
	mock.Stub("helloworld.Greeter", "SayHello").
		When(sdk.Matches("name", ".+")).
		Return("message", "Hi {{.Request.name}}").
		Delay(time.Duration(delayMs) * time.Millisecond).
		Commit()

	// Act
	reply, err := client.SayHello(t.Context(), &timed.HelloRequest{Name: "Alex"})

	// Assert
	require.NoError(t, err)
	require.Equal(t, "Hi Alex", reply.GetMessage())
	require.GreaterOrEqual(t, reply.GetDurationMs(), int64(delayMs))
}

func TestTimedGreeter_SayHello_WithDelay(t *testing.T) {
	t.Parallel()

	// Arrange
	client, mock := setupTimedServer(t)
	delayMs := 50
	mock.Stub("helloworld.Greeter", "SayHello").
		Unary("name", "Slow", "message", "Hello Slow").
		Delay(time.Duration(delayMs) * time.Millisecond).
		Commit()

	// Act
	reply, err := client.SayHello(t.Context(), &timed.HelloRequest{Name: "Slow"})

	// Assert
	require.NoError(t, err)
	require.Equal(t, "Hello Slow", reply.GetMessage())
	require.GreaterOrEqual(t, reply.GetDurationMs(), int64(delayMs))
}
