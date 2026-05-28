package admin

import (
	"io"
	"net/http"
	"os"
	"strings"
)

func LLMGatewayProxy(w http.ResponseWriter, r *http.Request) {
	llmGatewayURL := os.Getenv("LLM_GATEWAY_URL")
	if llmGatewayURL == "" {
		llmGatewayURL = "http://llm-gateway:39404"
	}

	var path string
	if strings.HasPrefix(r.URL.Path, "/llm-gateway") {
		path = r.URL.Path[len("/llm-gateway"):]
	} else {
		path = r.URL.Path
	}

	targetURL := llmGatewayURL + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
