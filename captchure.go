package captchure

import (
	"time"
	"errors"
	"fmt"
	"encoding/json"
	"net/url"
	"net/http"
	"bytes"
	"crypto/tls"
	"os"
	"bufio"
	"encoding/base64"
	"strconv"
)

type Captchure struct {
	ClientKey string 				`json:"clientKey"`
	Task map[string]interface{} 	`json:"task"`
	TaskId float64					`json:"taskId"`
	Solution map[string]interface{}	`json:"-"`
	Interval time.Duration			`json:"-"`
	Verbose bool					`json:"-"`
	Proxy string					`json:"-"`
	Trustful bool					`json:"-"`
}

func (s *Captchure) Publish() (err error) {
	if s.Verbose {
		fmt.Println("publishing a task...")
	}

	if s.ClientKey == "" {
		err = errors.New("empty clientKey")

		if s.Verbose {
			fmt.Println(err)
		}

		return err
	}

	if len(s.Task) == 0 {
		err = errors.New("empty task")

		if s.Verbose {
			fmt.Println(err)
		}

		return err
	}

	body, err := json.Marshal(s)

	if err != nil {
		err = errors.New("unable to marshall given task: " + err.Error())

		if s.Verbose {
			fmt.Println(err)
		}

		return err
	}

	var tempInterface interface{}

	json.Unmarshal(body, &tempInterface)

	tempStruct := tempInterface.(map[string]interface{})

	delete(tempStruct, "taskId")

	body, _ = json.Marshal(tempStruct)

	content, err := s.communicateWithService("/createTask", body)

	if err != nil {
		err = errors.New("service communication error: " + err.Error())

		if s.Verbose {
			fmt.Println(err)
		}

		return err
	}

	_, ok := content["taskId"].(float64)

	if !ok {
		errorCode, ok := content["errorCode"].(string)

		if !ok {
			errorCode = "unknown"
		}

		err = errors.New("unable to get result of a task creation: " + errorCode)

		if s.Verbose {
			fmt.Println(err)
		}

		return err
	}

	s.TaskId = content["taskId"].(float64)

	if s.Verbose {
		fmt.Println("task published: " + strconv.FormatFloat(s.TaskId, 'f', 0, 64))
	}

	return nil
}

func (s *Captchure) GetSolution() (err error) {
	if s.Verbose {
		fmt.Println("trying to get a solution...")
	}

	if s.ClientKey == "" {
		err = errors.New("empty clientKey")

		if s.Verbose {
			fmt.Println(err)
		}

		return err
	}

	if s.TaskId == 0 {
		err = errors.New("empty taskId")

		if s.Verbose {
			fmt.Println(err)
		}

		return err
	}

	body, err := json.Marshal(s)

	if err != nil {
		err = errors.New("unable to marshall given task: " + err.Error())

		if s.Verbose {
			fmt.Println(err)
		}

		return err
	}

	var tempInterface interface{}

	json.Unmarshal(body, &tempInterface)

	tempStruct := tempInterface.(map[string]interface{})

	delete(tempStruct, "task")

	body, _ = json.Marshal(tempStruct)

	if s.Interval == 0 {
		s.Interval = 1 * time.Second
	}

	for {
		if s.Verbose {
			fmt.Println("waiting for a solution...")
		}

		content, err := s.communicateWithService("/getTaskResult", body)

		if err != nil {
			err = errors.New("service communication error: " + err.Error())

			if s.Verbose {
				fmt.Println(err)
			}

			return err
		}

		_, ok := content["status"].(string)

		if !ok {
			errorCode, ok := content["errorCode"].(string)

			if !ok {
				errorCode = "unknown"
			}

			err = errors.New("unable to get status of a task: " + errorCode)

			if s.Verbose {
				fmt.Println(err)
			}

			return err
		}

		if content["status"].(string) != "processing" {
			_, ok := content["solution"].(map[string]interface{})

			if !ok {
				errorCode, ok := content["errorCode"].(string)

				if !ok {
					errorCode = "unknown"
				}

				err = errors.New("unable to get solution of a task: " + errorCode)

				if s.Verbose {
					fmt.Println(err)
				}

				return err
			}

			_, ok = content["solution"].(map[string]interface{})

			if !ok {
				errorCode, ok := content["errorCode"].(string)

				if !ok {
					errorCode = "unknown"
				}

				err = errors.New("unable to get a text of task solution: " + errorCode)

				if s.Verbose {
					fmt.Println(err)
				}

				return err
			}

			s.Solution = content["solution"].(map[string]interface{})

			return nil
		}

		time.Sleep(s.Interval)
	}
}

func (s *Captchure) SolveImage(base64image string, task map[string]interface{}) (word string, err error) {
	_, err = base64.StdEncoding.DecodeString(base64image)

	if err != nil {
		return word, errors.New("unable to validate given base64image")
	}

	task["type"] = "ImageToTextTask"
	task["body"] = base64image

	s.Task = task

	err = s.Publish()

	if err != nil {
		return word, err
	}

	err = s.GetSolution()

	if err != nil {
		return word, err
	}

	word, ok := s.Solution["text"].(string)

	if !ok {
		err = errors.New("unable to find target field in the solution")

		if s.Verbose {
			fmt.Println(err)
		}

		return word, err
	}

	if s.Verbose {
		fmt.Println("solution got: " + word)
	}

	return word, nil
}

func (s *Captchure) SolveRecaptcha(websiteUrl string, websiteKey string, task map[string]interface{}) (word string, err error) {
	task["websiteURL"] = websiteUrl
	task["websiteKey"] = websiteKey

	_, ok := task["type"]

	if !ok {
		task["type"] = "NoCaptchaTaskProxyless"
	}

	s.Task = task

	err = s.Publish()

	if err != nil {
		return word, err
	}

	err = s.GetSolution()

	if err != nil {
		return word, err
	}

	word, ok = s.Solution["gRecaptchaResponse"].(string)

	if !ok {
		err = errors.New("unable to find target field in the solution")

		if s.Verbose {
			fmt.Println(err)
		}

		return word, err
	}

	if s.Verbose {
		fmt.Println("solution got: " + word)
	}

	return word, nil
}

func (s *Captchure) SolveFunCaptcha(websiteUrl string, websiteKey string, task map[string]interface{}) (word string, err error) {
	task["type"] = "FunCaptchaTask"
	task["websiteURL"] = websiteUrl
	task["websiteKey"] = websiteKey

	s.Task = task

	err = s.Publish()

	if err != nil {
		return word, err
	}

	err = s.GetSolution()

	if err != nil {
		return word, err
	}

	word, ok := s.Solution["token"].(string)

	if !ok {
		err = errors.New("unable to find target field in the solution")

		if s.Verbose {
			fmt.Println(err)
		}

		return word, err
	}

	if s.Verbose {
		fmt.Println("solution got: " + word)
	}

	return word, nil
}

func (s *Captchure) communicateWithService(endpoint string, body []byte) (content map[string]interface{}, err error) {
	client := &http.Client{}

	transport := http.Transport{}

	if s.Proxy != "" {
		proxyHost, err := url.Parse(s.Proxy)

		if err != nil {
			return content, err
		}

		transport.Proxy = http.ProxyURL(proxyHost)
	}

	if s.Trustful {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client.Transport = &transport

	uri := url.URL{Scheme: "https", Host: "api.anti-captcha.com", Path: endpoint}

	response, err := client.Post(uri.String(), "application/json", bytes.NewBuffer(body))

	if err != nil {
		return content, err
	}

	defer response.Body.Close()

	content = make(map[string]interface{})

	err = json.NewDecoder(response.Body).Decode(&content)

	if err != nil {
		return content, err
	}

	return content, err
}

func LocalFileToBase64(path string) (encoded string, err error) {
	file, err := os.Open(path)

	if err != nil {
		return encoded, err
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	buffer := make([]byte, fileSize)

	fileReader := bufio.NewReader(file)
	fileReader.Read(buffer)

	return base64.StdEncoding.EncodeToString(buffer), nil
}