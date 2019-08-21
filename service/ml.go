package main

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/pkg/errors"
  "golang.org/x/oauth2/google"
  "io"
  "io/ioutil"
  "net/http"
  "strings"
)

type Prediction struct {
  Prediction int       `json:"prediction"`
  Key        string    `json:"key"`
  Scores     []float64 `json:"scores"`
}

type MlResponse struct {
  Predictions []Prediction `json:"predictions"`
}

type ImageBytes struct {
  B64 []byte `json:"b64"`
}
type Instance struct {
  ImageBytes ImageBytes `json:"image_bytes"`
  Key        string     `json:"key"`
}

type MlRequest struct {
  Instances []Instance `json:"instances"`
}

var (
  // TODO: Replace this project name and model name with your ML configuration.
  project = "around-faceml"
  model = "face_abc"
  url   = "https://ml.googleapis.com/v1/projects/" + project + "/models/" + model + ":predict"
  scope = "https://www.googleapis.com/auth/cloud-platform"
)

// Annotate a image file based on ml model, return score and error if exists.
func annotate(r io.Reader) (float64, error) {
	ctx := context.Background()
	buf, _ := ioutil.ReadAll(r)

	ts, err := google.DefaultTokenSource(ctx, scope)
	if err != nil {
	   fmt.Printf("failed to create token %v\n", err)
	   return 0.0, err
	}
	tt, _ := ts.Token()

	// Step 1, prepare request to Google ML API
	// Construct a ml request.
	request := &MlRequest{
	   Instances: []Instance{
		  {
			 ImageBytes: ImageBytes{
				B64: buf,
			 },
			 Key: "1", // Does not matter to the client, it's for Google tracking.
		  },
	   },
	}
	body, _ := json.Marshal(request)

	// Step 2, send the request
	// Construct a http request.
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+tt.AccessToken)

	fmt.Printf("Sending request to ml engine for prediction %s with token as %s\n", url, tt.AccessToken)
	// Send request to Google.
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
	   fmt.Printf("failed to send ml request %v\n", err)
	   return 0.0, err
	}

	// Step 3 parse the response
	var resp MlResponse
	body, _ = ioutil.ReadAll(res.Body)

	// Double check if the response is empty. Sometimes Google does not return an error instead just an
	// empty response while usually it's due to auth.
	if len(body) == 0 {
	   fmt.Println("empty google response")
	   return 0.0, errors.New("empty google response")
	}
	if err := json.Unmarshal(body, &resp); err != nil {
	   fmt.Printf("failed to parse response %v\n", err)
	   return 0.0, err
	}

	if len(resp.Predictions) == 0 {
	   // If the response is not empty, Google returns a different format. Check the raw message.
	   // Sometimes it's due to the image format. Google only accepts jpeg don't send png or others.
	   fmt.Printf("failed to parse response %s\n", string(body))
	   return 0.0, errors.Errorf("cannot parse response %s\n", string(body))
	}

	// Step 4, get the score
	// TODO: update index based on your ml model.
	results := resp.Predictions[0]
	fmt.Printf("Received a prediction result %f\n", results.Scores[0])

	// Step 5, return score
	return results.Scores[0], nil
  }