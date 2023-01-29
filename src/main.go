package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

func newsimpleServer(addr string) *simpleServer {

	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func newLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error : %v\n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string { return s.addr }

func (s *simpleServer) IsAlive() bool { return true }

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]

	}
	lb.roundRobinCount++
	return server
}

func (lb *LoadBalancer) serverProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("forwarding request to address %q\n", targetServer.Address())
	targetServer.Serve(rw, req)
}

func main() {
	servers := []Server{
		newsimpleServer("https://www.facebook.com"),
		newsimpleServer("http://www.bing.con"),
		newsimpleServer("http://www.duckduckgo.com"),
	}
	lb := newLoadBalancer("8000", servers)
	handleDirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serverProxy(rw, req)
	}
	http.HandleFunc("/", handleDirect)

	fmt.Printf("servering request at 'localhost:%s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
