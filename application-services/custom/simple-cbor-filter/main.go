//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
)

const (
	serviceKey = "sampleCborFilter"
)

var counter int = 0

func main() {
	// turn off secure mode for examples. Not recommended for production
	os.Setenv("EDGEX_SECURITY_SECRET_STORE", "false")
	// 1) First thing to do is to create an instance of the EdgeX SDK and initialize it.
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	// 2) shows how to access the application's specific configuration settings.
	valueDescriptors, err := edgexSdk.GetAppSettingStrings("ValueDescriptors")
	if err != nil {
		edgexSdk.LoggingClient.Error(err.Error())
		os.Exit(-1)
	}
	edgexSdk.LoggingClient.Info(fmt.Sprintf("Filtering for ValueDescriptors %v", valueDescriptors))

	// 3) This is our pipeline configuration, the collection of functions to
	// execute every time an event is triggered.
	edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(valueDescriptors).FilterByValueDescriptor,
		processImages,
	)

	// 4) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}

func processImages(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	event, ok := params[0].(models.Event)
	if !ok {
		return false, errors.New("processImages didn't receive expect models.Event type")

	}

	for _, reading := range event.Readings {
		// For this to work the image/jpeg & image/png packages must be imported to register their decoder
		imageData, imageType, err := image.Decode(bytes.NewReader(reading.BinaryValue))

		if err != nil {
			return false, errors.New("unable to decode image: " + err.Error())
		}

		// Since this is a example, we will just print put some stats from the images received
		fmt.Printf("Received Image from Device: %s, ReadingName: %s, Image Type: %s, Image Size: %s, Color in middle: %v\n",
			reading.Device, reading.Name, imageType, imageData.Bounds().Size().String(),
			imageData.At(imageData.Bounds().Size().X/2, imageData.Bounds().Size().Y/2))
	}

	return false, nil
}
