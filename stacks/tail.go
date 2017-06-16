package stacks

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// tail - tracks the progress during stack updates. c - command Type
func (s *Stack) tail(c string, done <-chan bool) {
	svc := cloudformation.New(s.Session, &aws.Config{Credentials: s.creds()})

	params := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(s.Stackname),
	}

	// used to track what lines have already been printed, to prevent dubplicate output
	printed := make(map[string]interface{})

	// create a ticker - 1.5 seconds
	tick := time.NewTicker(time.Millisecond * 1500)
	defer tick.Stop()

	for _ = range tick.C {
		select {
		case <-done:
			Log.Debug("Tail run.Completed")
			return
		default:
			// If channel is not populated, run verbose cf print
			Log.Debug(fmt.Sprintf("Calling [DescribeStackEvents] with parameters: %s", params))
			stackevents, err := svc.DescribeStackEvents(params)
			if err != nil {
				Log.Debug(fmt.Sprintln("Error when tailing events: ", err.Error()))
				continue
			}

			Log.Debug(fmt.Sprintln("Response:", stackevents))

			for _, event := range stackevents.StackEvents {

				statusReason := ""
				if strings.Contains(*event.ResourceStatus, "FAILED") {
					statusReason = *event.ResourceStatusReason
				}

				line := strings.Join([]string{
					*event.StackName,
					Log.ColorMap(*event.ResourceStatus),
					*event.ResourceType,
					*event.LogicalResourceId,
					statusReason,
				}, " - ")

				if _, ok := printed[line]; !ok {
					event := strings.Split(*event.ResourceStatus, "_")[0]
					if event == c || c == "" || strings.Contains(strings.ToLower(event), "rollback") {
						Log.Info(strings.Trim(line, "- "))
					}

					printed[line] = nil
				}
			}
		}

	}
}
