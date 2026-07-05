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

func runGreeterMock(t *testing.T) (*sdk.Server, helloworld.GreeterClient) {
	t.Helper()
	srv := sdk.NewServer(t, sdk.WithFileDescriptor(helloworld.File_greeter_proto))

	return srv, helloworld.NewGreeterClient(srv.Conn())
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

func setupTimedServer(t *testing.T) (timed.TimedGreeterClient, *sdk.Server) {
	t.Helper()
	srv, greeterClient := runGreeterMock(t)

	return runTimedServer(t, greeterClient), srv
}

func TestTimedGreeterSayHello(t *testing.T) {
	t.Parallel()

	// Arrange
	client, srv := setupTimedServer(t)
	delayMs := 20
	srv.ExpectUnary(helloworld.Greeter_SayHello_FullMethodName).
		Match("name", "Bob").
		Return(sdk.Delay(time.Duration(delayMs)*time.Millisecond, "message", "Hello Bob"))

	// Act
	reply, err := client.SayHello(t.Context(), &timed.HelloRequest{Name: "Bob"})

	// Assert
	require.NoError(t, err)
	require.Equal(t, "Hello Bob", reply.GetMessage())
	require.GreaterOrEqual(t, reply.GetDurationMs(), int64(delayMs))
}

func TestTimedGreeterSayHelloDynamicTemplate(t *testing.T) {
	t.Parallel()

	// Arrange
	client, srv := setupTimedServer(t)
	delayMs := 30
	srv.ExpectUnary(helloworld.Greeter_SayHello_FullMethodName).
		Match(sdk.Matches("name", ".+")).
		Return(sdk.Delay(time.Duration(delayMs)*time.Millisecond, "message", "Hi {{.Request.name}}"))

	// Act
	reply, err := client.SayHello(t.Context(), &timed.HelloRequest{Name: "Alex"})

	// Assert
	require.NoError(t, err)
	require.Equal(t, "Hi Alex", reply.GetMessage())
	require.GreaterOrEqual(t, reply.GetDurationMs(), int64(delayMs))
}

func TestTimedGreeterSayHelloWithDelay(t *testing.T) {
	t.Parallel()

	// Arrange
	client, srv := setupTimedServer(t)
	delayMs := 50
	srv.ExpectUnary(helloworld.Greeter_SayHello_FullMethodName).
		Match("name", "Slow").
		Return(sdk.Delay(time.Duration(delayMs)*time.Millisecond, "message", "Hello Slow"))

	// Act
	reply, err := client.SayHello(t.Context(), &timed.HelloRequest{Name: "Slow"})

	// Assert
	require.NoError(t, err)
	require.Equal(t, "Hello Slow", reply.GetMessage())
	require.GreaterOrEqual(t, reply.GetDurationMs(), int64(delayMs))
}
