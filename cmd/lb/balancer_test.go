package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "gopkg.in/check.v1"
)

type LoadBalancerSuite struct{}

var _ = Suite(&LoadBalancerSuite{})

func Test(t *testing.T) {
	TestingT(t)
}

func (s *LoadBalancerSuite) SetUpSuite(c *C) {

}

func (s *LoadBalancerSuite) TestLoadBalancing(c *C) {
	var server1Hits, server2Hits, server3Hits int

	backend1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		server1Hits++
		rw.WriteHeader(http.StatusOK)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		server2Hits++
		rw.WriteHeader(http.StatusOK)
	}))
	defer backend2.Close()

	backend3 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		server3Hits++
		rw.WriteHeader(http.StatusOK)
	}))
	defer backend3.Close()

	serversPool = []*Server{
		{URL: backend1.Listener.Addr().String(), Healthy: true},
		{URL: backend2.Listener.Addr().String(), Healthy: true},
		{URL: backend3.Listener.Addr().String(), Healthy: true},
	}

	request := httptest.NewRequest("GET", "http://example.com/foo", nil)

	for i := 0; i < 30; i++ {
		tempReq := request.Clone(request.Context())
		tempReq.URL.Path = fmt.Sprintf("/%d", i)
		response := httptest.NewRecorder()
		forward(response, tempReq)
	}

	c.Assert(server1Hits > 0, Equals, true)
	c.Assert(server2Hits > 0, Equals, true)
	c.Assert(server3Hits > 0, Equals, true)

	//fmt.Printf("Server1 hits: %d\n", server1Hits)
	//fmt.Printf("Server2 hits: %d\n", server2Hits)
	//fmt.Printf("Server3 hits: %d\n", server3Hits)
}

func (s *LoadBalancerSuite) TestNoHealthyServers(c *C) {
	serversPool = []*Server{
		{URL: "invalid:8080", Healthy: false},
	}

	request := httptest.NewRequest("GET", "http://example.com/foo", nil)
	response := httptest.NewRecorder()

	err := forward(response, request)

	c.Assert(err, NotNil)
	c.Assert(response.Code, Equals, http.StatusServiceUnavailable)
}
