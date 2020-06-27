package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsElasticTranscoderPreset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticTranscoderPresetCreate,
		Read:   resourceAwsElasticTranscoderPresetRead,
		Delete: resourceAwsElasticTranscoderPresetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"audio": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// elastictranscoder.AudioParameters
					Schema: map[string]*schema.Schema{
						"audio_packing_mode": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"SingleTrack",
								"OneChannelPerTrack",
								"OneChannelPerTrackWithMosTo8Tracks",
							}, false),
						},
						"bit_rate": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"channels": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"auto",
								"0",
								"1",
								"2",
							}, false),
						},
						"codec": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"AAC",
								"flac",
								"mp2",
								"mp3",
								"pcm",
								"vorbis",
							}, false),
						},
						"sample_rate": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"auto",
								"22050",
								"32000",
								"44100",
								"48000",
								"96000",
							}, false),
						},
					},
				},
			},
			"audio_codec_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bit_depth": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"8",
								"16",
								"24",
								"32",
							}, false),
						},
						"bit_order": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"LittleEndian",
							}, false),
						},
						"profile": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"auto",
								"AAC-LC",
								"HE-AAC",
								"HE-AACv2",
							}, false),
						},
						"signed": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Signed",
								"Unsigned",
							}, false),
						},
					},
				},
			},

			"container": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"flac",
					"flv",
					"fmp4",
					"gif",
					"mp2",
					"mp3",
					"mp4",
					"mpg",
					"mxf",
					"oga",
					"ogg",
					"ts",
					"wav",
					"webm",
				}, false),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"thumbnails": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					// elastictranscoder.Thumbnails
					Schema: map[string]*schema.Schema{
						"aspect_ratio": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"auto",
								"1:1",
								"4:3",
								"3:2",
								"16:9",
							}, false),
						},
						"format": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"jpg",
								"png",
							}, false),
						},
						"interval": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_height": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_width": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"padding_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Pad",
								"NoPad",
							}, false),
						},
						"resolution": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"sizing_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Fit",
								"Fill",
								"Stretch",
								"Keep",
								"ShrinkToFit",
								"ShrinkToFill",
							}, false),
						},
					},
				},
			},

			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Custom",
					"System",
				}, false),
			},

			"video": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// elastictranscoder.VideoParameters
					Schema: map[string]*schema.Schema{
						"aspect_ratio": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"auto",
								"1:1",
								"4:3",
								"3:2",
								"16:9",
							}, false),
						},
						"bit_rate": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"codec": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"gif",
								"H.264",
								"mpeg2",
								"vp8",
								"vp9",
							}, false),
						},
						"display_aspect_ratio": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"auto",
								"1:1",
								"4:3",
								"3:2",
								"16:9",
							}, false),
						},
						"fixed_gop": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"true",
								"false",
							}, false),
						},
						"frame_rate": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"auto",
								"10",
								"15",
								"23.97",
								"24",
								"25",
								"29.97",
								"30",
								"50",
								"60",
							}, false),
						},
						"keyframes_max_dist": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_frame_rate": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"10",
								"15",
								"23.97",
								"24",
								"25",
								"29.97",
								"30",
								"50",
								"60",
							}, false),
						},
						"max_height": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_width": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"padding_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Pad",
								"NoPad",
							}, false),
						},
						"resolution": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"sizing_policy": {
							Type:     schema.TypeString,
							Default:  "Fit",
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Fit",
								"Fill",
								"Stretch",
								"Keep",
								"ShrinkToFit",
								"ShrinkToFill",
							}, false),
						},
					},
				},
			},

			"video_watermarks": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					// elastictranscoder.PresetWatermark
					Schema: map[string]*schema.Schema{
						"horizontal_align": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Left",
								"Right",
								"Center",
							}, false),
						},
						"horizontal_offset": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 40),
						},
						"max_height": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"max_width": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"opacity": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"sizing_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Fit",
								"Stretch",
								"ShrinkToFit",
							}, false),
						},
						"target": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Content",
								"Frame",
							}, false),
						},
						"vertical_align": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Top",
								"Bottom",
								"Center",
							}, false),
						},
						"vertical_offset": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"video_codec_options": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsElasticTranscoderPresetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elastictranscoderconn

	req := &elastictranscoder.CreatePresetInput{
		Audio:       expandETAudioParams(d),
		Container:   aws.String(d.Get("container").(string)),
		Description: aws.String(d.Get("description").(string)),
		Thumbnails:  expandETThumbnails(d),
		Video:       expandETVideoParams(d),
	}

	if name, ok := d.GetOk("name"); ok {
		req.Name = aws.String(name.(string))
	} else {
		name := resource.PrefixedUniqueId("tf-et-preset-")
		d.Set("name", name)
		req.Name = aws.String(name)
	}

	log.Printf("[DEBUG] Elastic Transcoder Preset create opts: %s", req)
	resp, err := conn.CreatePreset(req)
	if err != nil {
		return fmt.Errorf("Error creating Elastic Transcoder Preset: %s", err)
	}

	if resp.Warning != nil && aws.StringValue(resp.Warning) != "" {
		log.Printf("[WARN] Elastic Transcoder Preset: %s", aws.StringValue(resp.Warning))
	}

	d.SetId(aws.StringValue(resp.Preset.Id))
	d.Set("arn", resp.Preset.Arn)

	return nil
}

func expandETThumbnails(d *schema.ResourceData) *elastictranscoder.Thumbnails {
	list, ok := d.GetOk("thumbnails")
	if !ok {
		return nil
	}

	l := list.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	t := l[0].(map[string]interface{})

	thumbnails := &elastictranscoder.Thumbnails{}

	if v, ok := t["aspect_ratio"]; ok && v.(string) != "" {
		thumbnails.AspectRatio = aws.String(v.(string))
	}

	if v, ok := t["interval"]; ok && v.(string) != "" {
		thumbnails.Interval = aws.String(v.(string))
	}

	if v, ok := t["format"]; ok && v.(string) != "" {
		thumbnails.Format = aws.String(v.(string))
	}

	if v, ok := t["max_height"]; ok && v.(string) != "" {
		thumbnails.MaxHeight = aws.String(v.(string))
	}

	if v, ok := t["max_width"]; ok && v.(string) != "" {
		thumbnails.MaxWidth = aws.String(v.(string))
	}

	if v, ok := t["padding_policy"]; ok && v.(string) != "" {
		thumbnails.PaddingPolicy = aws.String(v.(string))
	}

	if v, ok := t["resolution"]; ok && v.(string) != "" {
		thumbnails.Resolution = aws.String(v.(string))
	}

	if v, ok := t["sizing_policy"]; ok && v.(string) != "" {
		thumbnails.SizingPolicy = aws.String(v.(string))
	}

	return thumbnails
}

func expandETAudioParams(d *schema.ResourceData) *elastictranscoder.AudioParameters {
	list, ok := d.GetOk("audio")
	if !ok {
		return nil
	}

	l := list.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	audio := l[0].(map[string]interface{})

	return &elastictranscoder.AudioParameters{
		AudioPackingMode: aws.String(audio["audio_packing_mode"].(string)),
		BitRate:          aws.String(audio["bit_rate"].(string)),
		Channels:         aws.String(audio["channels"].(string)),
		Codec:            aws.String(audio["codec"].(string)),
		CodecOptions:     expandETAudioCodecOptions(d),
		SampleRate:       aws.String(audio["sample_rate"].(string)),
	}
}

func expandETAudioCodecOptions(d *schema.ResourceData) *elastictranscoder.AudioCodecOptions {
	l := d.Get("audio_codec_options").([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	codec := l[0].(map[string]interface{})

	codecOpts := &elastictranscoder.AudioCodecOptions{}

	if v, ok := codec["signed"]; ok && v.(string) != "" {
		codecOpts.Signed = aws.String(v.(string))
	}

	if v, ok := codec["profile"]; ok && v.(string) != "" {
		codecOpts.Profile = aws.String(v.(string))
	}

	if v, ok := codec["bit_order"]; ok && v.(string) != "" {
		codecOpts.BitOrder = aws.String(v.(string))
	}

	if v, ok := codec["bit_depth"]; ok && v.(string) != "" {
		codecOpts.BitDepth = aws.String(v.(string))
	}

	return codecOpts
}

func expandETVideoParams(d *schema.ResourceData) *elastictranscoder.VideoParameters {
	l := d.Get("video").([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	p := l[0].(map[string]interface{})

	etVideoParams := &elastictranscoder.VideoParameters{
		Watermarks: expandETVideoWatermarks(d),
	}

	if v, ok := d.GetOk("video_codec_options"); ok {
		codecOpts := make(map[string]string)
		for k, va := range v.(map[string]interface{}) {
			codecOpts[k] = va.(string)
		}

		etVideoParams.CodecOptions = aws.StringMap(codecOpts)
	}

	if v, ok := p["aspect_ratio"]; ok && v.(string) != "" {
		etVideoParams.AspectRatio = aws.String(v.(string))
	}

	if v, ok := p["bit_rate"]; ok && v.(string) != "" {
		etVideoParams.BitRate = aws.String(v.(string))
	}

	if v, ok := p["display_aspect_ratio"]; ok && v.(string) != "" {
		etVideoParams.DisplayAspectRatio = aws.String(v.(string))
	}

	if v, ok := p["aspect_ratio"]; ok && v.(string) != "" {
		etVideoParams.AspectRatio = aws.String(v.(string))
	}

	if v, ok := p["fixed_gop"]; ok && v.(string) != "" {
		etVideoParams.FixedGOP = aws.String(v.(string))
	}

	if v, ok := p["frame_rate"]; ok && v.(string) != "" {
		etVideoParams.FrameRate = aws.String(v.(string))
	}

	if v, ok := p["keyframes_max_dist"]; ok && v.(string) != "" {
		etVideoParams.KeyframesMaxDist = aws.String(v.(string))
	}

	if v, ok := p["max_frame_rate"]; ok && v.(string) != "" {
		etVideoParams.MaxFrameRate = aws.String(v.(string))
	}

	if v, ok := p["max_height"]; ok && v.(string) != "" {
		etVideoParams.MaxHeight = aws.String(v.(string))
	}

	if v, ok := p["max_width"]; ok && v.(string) != "" {
		etVideoParams.MaxWidth = aws.String(v.(string))
	}

	if v, ok := p["padding_policy"]; ok && v.(string) != "" {
		etVideoParams.PaddingPolicy = aws.String(v.(string))
	}

	if v, ok := p["resolution"]; ok && v.(string) != "" {
		etVideoParams.Resolution = aws.String(v.(string))
	}

	if v, ok := p["sizing_policy"]; ok && v.(string) != "" {
		etVideoParams.SizingPolicy = aws.String(v.(string))
	}

	if v, ok := p["codec"]; ok && v.(string) != "" {
		etVideoParams.Codec = aws.String(v.(string))
	}

	return etVideoParams
}

func expandETVideoWatermarks(d *schema.ResourceData) []*elastictranscoder.PresetWatermark {
	s := d.Get("video_watermarks").(*schema.Set)
	if s == nil || s.Len() == 0 {
		return nil
	}
	var watermarks []*elastictranscoder.PresetWatermark

	for _, w := range s.List() {
		if w == nil {
			continue
		}

		p := w.(map[string]interface{})
		watermark := &elastictranscoder.PresetWatermark{
			HorizontalAlign:  aws.String(p["horizontal_align"].(string)),
			HorizontalOffset: aws.String(p["horizontal_offset"].(string)),
			Id:               aws.String(p["id"].(string)),
			MaxHeight:        aws.String(p["max_height"].(string)),
			MaxWidth:         aws.String(p["max_width"].(string)),
			Opacity:          aws.String(p["opacity"].(string)),
			SizingPolicy:     aws.String(p["sizing_policy"].(string)),
			Target:           aws.String(p["target"].(string)),
			VerticalAlign:    aws.String(p["vertical_align"].(string)),
			VerticalOffset:   aws.String(p["vertical_offset"].(string)),
		}
		watermarks = append(watermarks, watermark)
	}

	return watermarks
}

func resourceAwsElasticTranscoderPresetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elastictranscoderconn

	resp, err := conn.ReadPreset(&elastictranscoder.ReadPresetInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, elastictranscoder.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] ElasticTranscoder Preset (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Elastic Transcoder Preset Read response: %#v", resp)

	preset := resp.Preset
	d.Set("arn", preset.Arn)

	if preset.Audio != nil {
		err := d.Set("audio", flattenETAudioParameters(preset.Audio))
		if err != nil {
			return err
		}

		if preset.Audio.CodecOptions != nil {
			d.Set("audio_codec_options", flattenETAudioCodecOptions(preset.Audio.CodecOptions))
		}
	}

	d.Set("container", preset.Container)
	d.Set("name", preset.Name)
	d.Set("description", preset.Description)

	if preset.Thumbnails != nil {
		err := d.Set("thumbnails", flattenETThumbnails(preset.Thumbnails))
		if err != nil {
			return err
		}
	}

	d.Set("type", preset.Type)

	if preset.Video != nil {
		err := d.Set("video", flattenETVideoParams(preset.Video))
		if err != nil {
			return err
		}

		if preset.Video.CodecOptions != nil {
			d.Set("video_codec_options", aws.StringValueMap(preset.Video.CodecOptions))
		}

		if preset.Video.Watermarks != nil {
			d.Set("video_watermarks", flattenETWatermarks(preset.Video.Watermarks))
		}
	}

	return nil
}

func flattenETAudioParameters(audio *elastictranscoder.AudioParameters) []map[string]interface{} {
	if audio == nil {
		return nil
	}

	result := map[string]interface{}{
		"audio_packing_mode": aws.StringValue(audio.AudioPackingMode),
		"bit_rate":           aws.StringValue(audio.BitRate),
		"channels":           aws.StringValue(audio.Channels),
		"codec":              aws.StringValue(audio.Codec),
		"sample_rate":        aws.StringValue(audio.SampleRate),
	}

	return []map[string]interface{}{result}
}

func flattenETAudioCodecOptions(opts *elastictranscoder.AudioCodecOptions) []map[string]interface{} {
	if opts == nil {
		return nil
	}

	result := map[string]interface{}{
		"bit_depth": aws.StringValue(opts.BitDepth),
		"bit_order": aws.StringValue(opts.BitOrder),
		"profile":   aws.StringValue(opts.Profile),
		"signed":    aws.StringValue(opts.Signed),
	}

	return []map[string]interface{}{result}
}

func flattenETThumbnails(thumbs *elastictranscoder.Thumbnails) []map[string]interface{} {
	if thumbs == nil {
		return nil
	}

	result := map[string]interface{}{
		"aspect_ratio":   aws.StringValue(thumbs.AspectRatio),
		"format":         aws.StringValue(thumbs.Format),
		"interval":       aws.StringValue(thumbs.Interval),
		"max_height":     aws.StringValue(thumbs.MaxHeight),
		"max_width":      aws.StringValue(thumbs.MaxWidth),
		"padding_policy": aws.StringValue(thumbs.PaddingPolicy),
		"resolution":     aws.StringValue(thumbs.Resolution),
		"sizing_policy":  aws.StringValue(thumbs.SizingPolicy),
	}

	return []map[string]interface{}{result}
}

func flattenETVideoParams(video *elastictranscoder.VideoParameters) []map[string]interface{} {
	if video == nil {
		return nil
	}

	result := map[string]interface{}{
		"aspect_ratio":         aws.StringValue(video.AspectRatio),
		"bit_rate":             aws.StringValue(video.BitRate),
		"codec":                aws.StringValue(video.Codec),
		"display_aspect_ratio": aws.StringValue(video.DisplayAspectRatio),
		"fixed_gop":            aws.StringValue(video.FixedGOP),
		"frame_rate":           aws.StringValue(video.FrameRate),
		"keyframes_max_dist":   aws.StringValue(video.KeyframesMaxDist),
		"max_frame_rate":       aws.StringValue(video.MaxFrameRate),
		"max_height":           aws.StringValue(video.MaxHeight),
		"max_width":            aws.StringValue(video.MaxWidth),
		"padding_policy":       aws.StringValue(video.PaddingPolicy),
		"resolution":           aws.StringValue(video.Resolution),
		"sizing_policy":        aws.StringValue(video.SizingPolicy),
	}

	return []map[string]interface{}{result}
}

func flattenETWatermarks(watermarks []*elastictranscoder.PresetWatermark) []map[string]interface{} {
	var watermarkSet []map[string]interface{}

	for _, w := range watermarks {
		watermark := map[string]interface{}{
			"horizontal_align":  aws.StringValue(w.HorizontalAlign),
			"horizontal_offset": aws.StringValue(w.HorizontalOffset),
			"id":                aws.StringValue(w.Id),
			"max_height":        aws.StringValue(w.MaxHeight),
			"max_width":         aws.StringValue(w.MaxWidth),
			"opacity":           aws.StringValue(w.Opacity),
			"sizing_policy":     aws.StringValue(w.SizingPolicy),
			"target":            aws.StringValue(w.Target),
			"vertical_align":    aws.StringValue(w.VerticalAlign),
			"vertical_offset":   aws.StringValue(w.VerticalOffset),
		}

		watermarkSet = append(watermarkSet, watermark)
	}

	return watermarkSet
}

func resourceAwsElasticTranscoderPresetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elastictranscoderconn

	log.Printf("[DEBUG] Elastic Transcoder Delete Preset: %s", d.Id())
	_, err := conn.DeletePreset(&elastictranscoder.DeletePresetInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting Elastic Transcoder Preset: %s", err)
	}

	return nil
}
