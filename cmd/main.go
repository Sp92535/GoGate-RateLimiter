// main.go
package main

import "github.com/Sp92535/GoGate-RateLimiter/internal/proxy"

func main() {
	// call to start the registered proxies
	proxy.Run()
	// config := utils.NewConfiguration("config/config.yaml")
	// configJSON, err := json.MarshalIndent(config, "", "  ")
	// if err != nil {
	// 	log.Fatalf("Error formatting config: %v", err)
	// }

	// fmt.Println(string(configJSON))
}
