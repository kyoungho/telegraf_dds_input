/*****************************************************************************
*   (c) 2005-2015 Copyright, Real-Time Innovations.  All rights reserved.    *
*                                                                            *
* No duplications, whole or partial, manual or electronic, may be made       *
* without express written permission.  Any such copies, or revisions thereof,*
* must display this notice unaltered.                                        *
* This code contains trade secrets of Real-Time Innovations, Inc.            *
*                                                                            *
*****************************************************************************/

package dds_consumer

import (
	"github.com/influxdata/telegraf"
	//"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/rticommunity/rticonnextdds-connector-go"
	"log"
	"time"
)

type DDSConsumer struct {
	// XML configuration file path
	ConfigFilePath string `toml:"config_path"`
	// XML configuration name for DDS Participant
	ParticipantConfig string `toml:"participant_config"`
	// XML configuration names for DDS Readers
	ReaderConfig string `toml:"reader_config"`
        TagKeys []string `toml:"tag_keys"`

	// RTI Connext Connector entities
	connector *rti.Connector
	reader    *rti.Input

	// Telegraf entities
	parser *Parser
	acc    telegraf.Accumulator
}

// Default configurations
var sampleConfig = `
  ## XML configuration file path
  config_path = "example_configs/ShapeExample.xml"

  ## Configuration name for DDS Participant from a description in XML
  participant_config = "MyParticipantLibrary::Zero"

  ## Configuration name for DDS DataReader from a description in XML
  reader_config = "MySubscriber::MySquareReader"

  # Tag key is an array of keys that should be added as tags.
  tag_keys = ["color"]

  # Override the base name of the measurement
  name_override = "shapes"

  ## Data format to consume.
  data_format = "json"
`

func checkFatalError(err error) {
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
}

func checkError(err error) {
	if err != nil {
		log.Println("ERROR:", err)
	}
}

func (d *DDSConsumer) SampleConfig() string {
	return sampleConfig
}

func (d *DDSConsumer) Description() string {
	return "Read metrics from DDS"
}

//func (d *DDSConsumer) SetParser(parser parsers.Parser) {
	//d.parser =  parser
//}

func (d *DDSConsumer) Start(acc telegraf.Accumulator) error {
	// Disable logs
	//log.SetFlags(0)
	//log.SetOutput(ioutil.Discard)

	// Keep the Telegraf accumulator internally
	d.acc = acc

	var err error

	// Create a Connector entity
	d.connector, err = rti.NewConnector(d.ParticipantConfig, d.ConfigFilePath)
	checkFatalError(err)

	// Get a DDS reader
	d.reader, err = d.connector.GetInput(d.ReaderConfig)
	checkFatalError(err)

        // Set tag keys
	var parserConfig Config
	parserConfig.TagKeys = d.TagKeys

	d.parser, err = New(&Config{MetricName: "dds_consumer"})
	checkFatalError(err)

	// Start a thread for ingesting DDS
	go d.process()

	return nil
}

func (d *DDSConsumer) Stop() {
	d.connector.Delete()
}

// Take DDS samples from the DataReader and ingest them to Telegraf outputs
func (d *DDSConsumer) process() {
	for {
		d.connector.Wait(-1)
		d.reader.Take()
		numOfSamples, err := d.reader.Samples.GetLength()
		checkError(err)
		if err != nil {
			continue
		}

		for i := 0; i < numOfSamples; i++ {
			valid, err := d.reader.Infos.IsValid(i)
			checkError(err)
			if err != nil {
				continue
			}
			if valid {
				json, err := d.reader.Samples.GetJSON(i)
				checkError(err)
				ts, err := d.reader.Infos.GetSourceTimestamp(i)
				checkError(err)
				go func(json []byte) {
					// Parse the JSON object to metrics
					metrics, err := d.parser.Parse(json)
					checkError(err)

					// Iterate the metrics
					for _, metric := range metrics {
						//log.Println(metric.Fields()["x"])
						//Add a metric to an output plugin
						d.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), time.Unix(0, ts))
						//d.acc.AddMetric(metric)
					}
				}(json)
			}
		}
	}
}

func (d *DDSConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("dds_consumer", func() telegraf.Input {
		return &DDSConsumer{}
	})
}
