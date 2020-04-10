package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSElasticTranscoderPreset_basic(t *testing.T) {
	preset := &elastictranscoder.Preset{}
	name := "aws_elastictranscoder_preset.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSElasticTranscoder(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckElasticTranscoderPresetDestroy,
		Steps: []resource.TestStep{
			{
				Config: awsElasticTranscoderPresetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(name, preset),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: awsElasticTranscoderPresetConfig2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(name, preset),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: awsElasticTranscoderPresetConfig3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(name, preset),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticTranscoderPreset_disappears(t *testing.T) {
	preset := &elastictranscoder.Preset{}
	name := "aws_elastictranscoder_preset.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSElasticTranscoder(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckElasticTranscoderPresetDestroy,
		Steps: []resource.TestStep{
			{
				Config: awsElasticTranscoderPresetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(name, preset),
					testAccCheckElasticTranscoderPresetDisappears(preset),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckElasticTranscoderPresetExists(name string, preset *elastictranscoder.Preset) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).elastictranscoderconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Preset ID is set")
		}

		out, err := conn.ReadPreset(&elastictranscoder.ReadPresetInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*preset = *out.Preset

		return nil
	}
}

func testAccCheckElasticTranscoderPresetDisappears(preset *elastictranscoder.Preset) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).elastictranscoderconn
		_, err := conn.DeletePreset(&elastictranscoder.DeletePresetInput{
			Id: preset.Id,
		})

		return err
	}
}

func testAccCheckElasticTranscoderPresetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elastictranscoderconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastictranscoder_preset" {
			continue
		}

		out, err := conn.ReadPreset(&elastictranscoder.ReadPresetInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if out.Preset != nil && *out.Preset.Id == rs.Primary.ID {
				return fmt.Errorf("Elastic Transcoder Preset still exists")
			}
		}

		if !isAWSErr(err, elastictranscoder.ErrCodeResourceNotFoundException, "") {
			return fmt.Errorf("unexpected error: %s", err)
		}

	}
	return nil
}

func awsElasticTranscoderPresetConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container   = "mp4"
  description = "elastic transcoder preset test 1"
  name        = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 320
    channels           = 2
    codec              = "mp3"
    sample_rate        = 44100
  }
}
`, rName)
}

func awsElasticTranscoderPresetConfig2(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container   = "mp4"
  description = "elastic transcoder preset test 2"
  name        = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 128
    channels           = 2
    codec              = "AAC"
    sample_rate        = 48000
  }

  audio_codec_options {
    profile = "auto"
  }

  video {
    bit_rate             = "auto"
    codec                = "H.264"
    display_aspect_ratio = "16:9"
    fixed_gop            = "true"
    frame_rate           = "auto"
    keyframes_max_dist   = 90
    max_height           = 1080
    max_width            = 1920
    padding_policy       = "Pad"
    sizing_policy        = "Fit"
  }

  video_codec_options = {
    Profile                  = "main"
    Level                    = "4.1"
    MaxReferenceFrames       = 4
    InterlacedMode           = "Auto"
    ColorSpaceConversionMode = "None"
  }

  thumbnails {
    format         = "jpg"
    interval       = 5
    max_width      = 960
    max_height     = 540
    padding_policy = "Pad"
    sizing_policy  = "Fit"
  }
}
`, rName)
}

func awsElasticTranscoderPresetConfig3(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container   = "mp4"
  description = "elastic transcoder preset test 3"
  name        = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 96
    channels           = 2
    codec              = "AAC"
    sample_rate        = 44100
  }

  audio_codec_options {
    profile = "AAC-LC"
  }

  video {
    bit_rate             = "1600"
    codec                = "H.264"
    display_aspect_ratio = "16:9"
    fixed_gop            = "false"
    frame_rate           = "auto"
    max_frame_rate       = "60"
    keyframes_max_dist   = 240
    max_height           = "auto"
    max_width            = "auto"
    padding_policy       = "Pad"
    sizing_policy        = "Fit"
  }

  video_codec_options = {
    Profile                  = "main"
    Level                    = "2.2"
    MaxReferenceFrames       = 3
    InterlacedMode           = "Progressive"
    ColorSpaceConversionMode = "None"
  }

  video_watermarks {
    id                = "Terraform Test"
    max_width         = "20%%"
    max_height        = "20%%"
    sizing_policy     = "ShrinkToFit"
    horizontal_align  = "Right"
    horizontal_offset = "10px"
    vertical_align    = "Bottom"
    vertical_offset   = "10px"
    opacity           = "55.5"
    target            = "Content"
  }

  thumbnails {
    format         = "png"
    interval       = 120
    max_width      = "auto"
    max_height     = "auto"
    padding_policy = "Pad"
    sizing_policy  = "Fit"
  }
}
`, rName)
}
