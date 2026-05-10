package services

import (
	ctx "context"
	"metrics-server/internal/usecase/context"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CheckAddr is used to check if x-real-ip metadata is whitelisted.
func CheckAddr(app *context.AppContext) func(
	ctx ctx.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	return func(
		ctx ctx.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if app.Cfg.TrustedSubnet != "" {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				app.Log.Errorln("Missing metadata")
				return nil, status.Error(codes.Unauthenticated, "Missing metadata")
			}

			ipHeaders := md.Get("x-real-ip")
			if len(ipHeaders) == 0 {
				app.Log.Errorln("Missing x-real-ip")
				return nil, status.Error(codes.Unauthenticated, "Missing x-real-ip")
			}

			ipHeader := ipHeaders[0]
			ip := net.ParseIP(ipHeader)
			if ip == nil {
				app.Log.Errorln("Invalid x-real-ip", ipHeaders[0])
				return nil, status.Error(codes.Unauthenticated, "Invalid x-real-ip")
			}

			whitelist := strings.Split(app.Cfg.TrustedSubnet, ",")
			var whitelisted bool
			for _, cidr := range whitelist {
				_, network, err := net.ParseCIDR(cidr)
				if err != nil {
					app.Log.Fatalln("Invalid network in whitelist:", cidr)
					return nil, status.Error(codes.Unauthenticated, "Internal error while checking IP")
				}
				if network.Contains(ip) {
					whitelisted = true
					break
				}
			}
			if !whitelisted {
				app.Log.Errorf("x-real-ip %s is not in whitelist", ip)
				return nil, status.Error(codes.Unauthenticated, "Forbidden")
			}
		}
		return handler(ctx, req)
	}
}
