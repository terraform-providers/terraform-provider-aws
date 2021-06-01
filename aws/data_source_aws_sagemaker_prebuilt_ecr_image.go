package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	// SageMaker Algorithm BlazingText
	sageMakerRepositoryBlazingText = "blazingtext"
	// SageMaker Algorithm DeepAR Forecasting
	sageMakerRepositoryDeepARForecasting = "forecasting-deepar"
	// SageMaker Algorithm Factorization Machines
	sageMakerRepositoryFactorizationMachines = "factorization-machines"
	// SageMaker Algorithm Image Classification
	sageMakerRepositoryImageClassification = "image-classification"
	// SageMaker Algorithm IP Insights
	sageMakerRepositoryIPInsights = "ipinsights"
	// SageMaker Algorithm k-means
	sageMakerRepositoryKMeans = "kmeans"
	// SageMaker Algorithm k-nearest-neighbor
	sageMakerRepositoryKNearestNeighbor = "knn"
	// SageMaker Algorithm Latent Dirichlet Allocation
	sageMakerRepositoryLDA = "lda"
	// SageMaker Algorithm Linear Learner
	sageMakerRepositoryLinearLearner = "linear-learner"
	// SageMaker Algorithm Neural Topic Model
	sageMakerRepositoryNeuralTopicModel = "ntm"
	// SageMaker Algorithm Object2Vec
	sageMakerRepositoryObject2Vec = "object2vec"
	// SageMaker Algorithm Object Detection
	sageMakerRepositoryObjectDetection = "object-detection"
	// SageMaker Algorithm PCA
	sageMakerRepositoryPCA = "pca"
	// SageMaker Algorithm Random Cut Forest
	sageMakerRepositoryRandomCutForest = "randomcutforest"
	// SageMaker Algorithm Semantic Segmentation
	sageMakerRepositorySemanticSegmentation = "semantic-segmentation"
	// SageMaker Algorithm Seq2Seq
	sageMakerRepositorySeq2Seq = "seq2seq"
	// SageMaker Algorithm XGBoost
	sageMakerRepositoryXGBoost = "sagemaker-xgboost"
	// SageMaker Library scikit-learn
	sageMakerRepositoryScikitLearn = "sagemaker-scikit-learn"
	// SageMaker Library Spark ML
	sageMakerRepositorySparkML = "sagemaker-sparkml-serving"
	// SageMaker Library TensorFlow Serving
	sageMakerRepositoryTensorFlowServing = "sagemaker-tensorflow-serving"
	// SageMaker Library TensorFlow Serving EIA
	sageMakerRepositoryTensorFlowServingEIA = "sagemaker-tensorflow-serving-eia"
	// SageMaker Repo MXNet Inference
	sageMakerRepositoryMXNetInference = "mxnet-inference"
	// SageMaker Repo MXNet Inference EIA
	sageMakerRepositoryMXNetInferenceEIA = "mxnet-inference-eia"
	// SageMaker Repo MXNet Training
	sageMakerRepositoryMXNetTraining = "mxnet-training"
	// SageMaker Repo PyTorch Inference
	sageMakerRepositoryPyTorchInference = "pytorch-inference"
	// SageMaker Repo PyTorch Inference EIA
	sageMakerRepositoryPyTorchInferenceEIA = "pytorch-inference-eia"
	// SageMaker Repo PyTorch Training
	sageMakerRepositoryPyTorchTraining = "pytorch-training"
	// SageMaker Repo TensorFlow Inference
	sageMakerRepositoryTensorFlowInference = "tensorflow-inference"
	// SageMaker Repo TensorFlow Inference EIA
	sageMakerRepositoryTensorFlowInferenceEIA = "tensorflow-inference-eia"
	// SageMaker Repo TensorFlow Training
	sageMakerRepositoryTensorFlowTraining = "tensorflow-training"
)

// https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
var sageMakerPrebuiltECRImageIDByRegion_Blazing = map[string]string{
	endpoints.ApEast1RegionID:      "286214385809",
	endpoints.ApNortheast1RegionID: "501404015308",
	endpoints.ApNortheast2RegionID: "306986355934",
	endpoints.ApSouth1RegionID:     "991648021394",
	endpoints.ApSoutheast1RegionID: "475088953585",
	endpoints.ApSoutheast2RegionID: "544295431143",
	endpoints.CaCentral1RegionID:   "469771592824",
	endpoints.CnNorth1RegionID:     "390948362332",
	endpoints.CnNorthwest1RegionID: "387376663083",
	endpoints.EuCentral1RegionID:   "813361260812",
	endpoints.EuNorth1RegionID:     "669576153137",
	endpoints.EuWest1RegionID:      "685385470294",
	endpoints.EuWest2RegionID:      "644912444149",
	endpoints.EuWest3RegionID:      "749696950732",
	endpoints.MeSouth1RegionID:     "249704162688",
	endpoints.SaEast1RegionID:      "855470959533",
	endpoints.UsEast1RegionID:      "811284229777",
	endpoints.UsEast2RegionID:      "825641698319",
	endpoints.UsGovWest1RegionID:   "226302683700",
	endpoints.UsWest1RegionID:      "632365934929",
	endpoints.UsWest2RegionID:      "433757028032",
}

// https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
var sageMakerPrebuiltECRImageIDByRegion_DeepAR = map[string]string{
	endpoints.ApEast1RegionID:      "286214385809",
	endpoints.ApNortheast1RegionID: "633353088612",
	endpoints.ApNortheast2RegionID: "204372634319",
	endpoints.ApSouth1RegionID:     "991648021394",
	endpoints.ApSoutheast1RegionID: "475088953585",
	endpoints.ApSoutheast2RegionID: "514117268639",
	endpoints.CaCentral1RegionID:   "469771592824",
	endpoints.CnNorth1RegionID:     "390948362332",
	endpoints.CnNorthwest1RegionID: "387376663083",
	endpoints.EuCentral1RegionID:   "495149712605",
	endpoints.EuNorth1RegionID:     "669576153137",
	endpoints.EuWest1RegionID:      "224300973850",
	endpoints.EuWest2RegionID:      "644912444149",
	endpoints.EuWest3RegionID:      "749696950732",
	endpoints.MeSouth1RegionID:     "249704162688",
	endpoints.SaEast1RegionID:      "855470959533",
	endpoints.UsEast1RegionID:      "522234722520",
	endpoints.UsEast2RegionID:      "566113047672",
	endpoints.UsGovWest1RegionID:   "226302683700",
	endpoints.UsWest1RegionID:      "632365934929",
	endpoints.UsWest2RegionID:      "156387875391",
}

// https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
var sageMakerPrebuiltECRImageIDByRegion_FactorMachines = map[string]string{
	endpoints.ApEast1RegionID:      "286214385809",
	endpoints.ApNortheast1RegionID: "351501993468",
	endpoints.ApNortheast2RegionID: "835164637446",
	endpoints.ApSouth1RegionID:     "991648021394",
	endpoints.ApSoutheast1RegionID: "475088953585",
	endpoints.ApSoutheast2RegionID: "712309505854",
	endpoints.CaCentral1RegionID:   "469771592824",
	endpoints.CnNorth1RegionID:     "390948362332",
	endpoints.CnNorthwest1RegionID: "387376663083",
	endpoints.EuCentral1RegionID:   "664544806723",
	endpoints.EuNorth1RegionID:     "669576153137",
	endpoints.EuWest1RegionID:      "438346466558",
	endpoints.EuWest2RegionID:      "644912444149",
	endpoints.EuWest3RegionID:      "749696950732",
	endpoints.MeSouth1RegionID:     "249704162688",
	endpoints.SaEast1RegionID:      "855470959533",
	endpoints.UsEast1RegionID:      "382416733822",
	endpoints.UsEast2RegionID:      "404615174143",
	endpoints.UsGovWest1RegionID:   "226302683700",
	endpoints.UsWest1RegionID:      "632365934929",
	endpoints.UsWest2RegionID:      "174872318107",
}

// https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
var sageMakerPrebuiltECRImageIDByRegion_LDA = map[string]string{
	endpoints.ApNortheast1RegionID: "258307448986",
	endpoints.ApNortheast2RegionID: "293181348795",
	endpoints.ApSouth1RegionID:     "991648021394",
	endpoints.ApSoutheast1RegionID: "475088953585",
	endpoints.ApSoutheast2RegionID: "297031611018",
	endpoints.CaCentral1RegionID:   "469771592824",
	endpoints.EuCentral1RegionID:   "353608530281",
	endpoints.EuWest1RegionID:      "999678624901",
	endpoints.EuWest2RegionID:      "644912444149",
	endpoints.UsEast1RegionID:      "766337827248",
	endpoints.UsEast2RegionID:      "999911452149",
	endpoints.UsGovWest1RegionID:   "226302683700",
	endpoints.UsWest1RegionID:      "632365934929",
	endpoints.UsWest2RegionID:      "266724342769",
}

// https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
var sageMakerPrebuiltECRImageIDByRegion_XGBoost = map[string]string{
	endpoints.ApEast1RegionID:      "651117190479",
	endpoints.ApNortheast1RegionID: "354813040037",
	endpoints.ApNortheast2RegionID: "366743142698",
	endpoints.ApSouth1RegionID:     "720646828776",
	endpoints.ApSoutheast1RegionID: "121021644041",
	endpoints.ApSoutheast2RegionID: "783357654285",
	endpoints.CaCentral1RegionID:   "341280168497",
	endpoints.CnNorth1RegionID:     "450853457545",
	endpoints.CnNorthwest1RegionID: "451049120500",
	endpoints.EuCentral1RegionID:   "492215442770",
	endpoints.EuNorth1RegionID:     "662702820516",
	endpoints.EuWest1RegionID:      "141502667606",
	endpoints.EuWest2RegionID:      "764974769150",
	endpoints.EuWest3RegionID:      "659782779980",
	endpoints.MeSouth1RegionID:     "801668240914",
	endpoints.SaEast1RegionID:      "737474898029",
	endpoints.UsEast1RegionID:      "683313688378",
	endpoints.UsEast2RegionID:      "257758044811",
	endpoints.UsGovWest1RegionID:   "414596584902",
	endpoints.UsWest1RegionID:      "746614075791",
	endpoints.UsWest2RegionID:      "246618743249",
}

// https://docs.aws.amazon.com/sagemaker/latest/dg/pre-built-docker-containers-scikit-learn-spark.html
var sageMakerPrebuiltECRImageIDByRegion_SparkML = map[string]string{
	endpoints.ApNortheast1RegionID: "354813040037",
	endpoints.ApNortheast2RegionID: "366743142698",
	endpoints.ApSouth1RegionID:     "720646828776",
	endpoints.ApSoutheast1RegionID: "121021644041",
	endpoints.ApSoutheast2RegionID: "783357654285",
	endpoints.CaCentral1RegionID:   "341280168497",
	endpoints.EuCentral1RegionID:   "492215442770",
	endpoints.EuWest1RegionID:      "141502667606",
	endpoints.EuWest2RegionID:      "764974769150",
	endpoints.UsEast1RegionID:      "683313688378",
	endpoints.UsEast2RegionID:      "257758044811",
	endpoints.UsGovWest1RegionID:   "414596584902",
	endpoints.UsWest1RegionID:      "746614075791",
	endpoints.UsWest2RegionID:      "246618743249",
}

// https://github.com/aws/deep-learning-containers/blob/master/available_images.md
// https://github.com/aws/sagemaker-tensorflow-serving-container
var sageMakerPrebuiltECRImageIDByRegion_DeepLearning = map[string]string{
	endpoints.ApEast1RegionID:      "871362719292",
	endpoints.ApNortheast1RegionID: "763104351884",
	endpoints.ApNortheast2RegionID: "763104351884",
	endpoints.ApSouth1RegionID:     "763104351884",
	endpoints.ApSoutheast1RegionID: "763104351884",
	endpoints.ApSoutheast2RegionID: "763104351884",
	endpoints.CaCentral1RegionID:   "763104351884",
	endpoints.CnNorth1RegionID:     "727897471807",
	endpoints.CnNorthwest1RegionID: "727897471807",
	endpoints.EuCentral1RegionID:   "763104351884",
	endpoints.EuNorth1RegionID:     "763104351884",
	endpoints.EuWest1RegionID:      "763104351884",
	endpoints.EuWest2RegionID:      "763104351884",
	endpoints.EuWest3RegionID:      "763104351884",
	endpoints.MeSouth1RegionID:     "217643126080",
	endpoints.SaEast1RegionID:      "763104351884",
	endpoints.UsEast1RegionID:      "763104351884",
	endpoints.UsEast2RegionID:      "763104351884",
	endpoints.UsIsoEast1RegionID:   "886529160074",
	endpoints.UsWest1RegionID:      "763104351884",
	endpoints.UsWest2RegionID:      "763104351884",
}

// https://github.com/aws/sagemaker-tensorflow-serving-container
var sageMakerPrebuiltECRImageIDByRegion_TensorFlowServing = map[string]string{
	endpoints.ApEast1RegionID:      "057415533634",
	endpoints.ApNortheast1RegionID: "520713654638",
	endpoints.ApNortheast2RegionID: "520713654638",
	endpoints.ApSouth1RegionID:     "520713654638",
	endpoints.ApSoutheast1RegionID: "520713654638",
	endpoints.ApSoutheast2RegionID: "520713654638",
	endpoints.CaCentral1RegionID:   "520713654638",
	endpoints.CnNorth1RegionID:     "520713654638",
	endpoints.CnNorthwest1RegionID: "520713654638",
	endpoints.EuCentral1RegionID:   "520713654638",
	endpoints.EuNorth1RegionID:     "520713654638",
	endpoints.EuWest1RegionID:      "520713654638",
	endpoints.EuWest2RegionID:      "520713654638",
	endpoints.EuWest3RegionID:      "520713654638",
	endpoints.MeSouth1RegionID:     "724002660598",
	endpoints.SaEast1RegionID:      "520713654638",
	endpoints.UsEast1RegionID:      "520713654638",
	endpoints.UsEast2RegionID:      "520713654638",
	endpoints.UsWest1RegionID:      "520713654638",
	endpoints.UsWest2RegionID:      "520713654638",
}

func dataSourceAwsSageMakerPrebuiltECRImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSageMakerPrebuiltECRImageRead,
		Schema: map[string]*schema.Schema{
			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					sageMakerRepositoryBlazingText,
					sageMakerRepositoryDeepARForecasting,
					sageMakerRepositoryFactorizationMachines,
					sageMakerRepositoryImageClassification,
					sageMakerRepositoryIPInsights,
					sageMakerRepositoryKMeans,
					sageMakerRepositoryKNearestNeighbor,
					sageMakerRepositoryLDA,
					sageMakerRepositoryLinearLearner,
					sageMakerRepositoryMXNetInference,
					sageMakerRepositoryMXNetInferenceEIA,
					sageMakerRepositoryMXNetTraining,
					sageMakerRepositoryNeuralTopicModel,
					sageMakerRepositoryObject2Vec,
					sageMakerRepositoryObjectDetection,
					sageMakerRepositoryPCA,
					sageMakerRepositoryPyTorchInference,
					sageMakerRepositoryPyTorchInferenceEIA,
					sageMakerRepositoryPyTorchTraining,
					sageMakerRepositoryRandomCutForest,
					sageMakerRepositoryScikitLearn,
					sageMakerRepositorySemanticSegmentation,
					sageMakerRepositorySeq2Seq,
					sageMakerRepositorySparkML,
					sageMakerRepositoryTensorFlowInference,
					sageMakerRepositoryTensorFlowInferenceEIA,
					sageMakerRepositoryTensorFlowServing,
					sageMakerRepositoryTensorFlowServingEIA,
					sageMakerRepositoryTensorFlowTraining,
					sageMakerRepositoryXGBoost,
				}, false),
			},

			"dns_suffix": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"image_tag": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1",
			},

			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"registry_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsSageMakerPrebuiltECRImageRead(d *schema.ResourceData, meta interface{}) error {
	region := meta.(*AWSClient).region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}

	suffix := meta.(*AWSClient).dnsSuffix
	if v, ok := d.GetOk("dns_suffix"); ok {
		suffix = v.(string)
	}

	repo := d.Get("repository_name").(string)

	id := ""
	switch repo {
	case sageMakerRepositoryBlazingText,
		sageMakerRepositoryImageClassification,
		sageMakerRepositoryObjectDetection,
		sageMakerRepositorySemanticSegmentation,
		sageMakerRepositorySeq2Seq:
		id = sageMakerPrebuiltECRImageIDByRegion_Blazing[region]
	case sageMakerRepositoryDeepARForecasting:
		id = sageMakerPrebuiltECRImageIDByRegion_DeepAR[region]
	case sageMakerRepositoryLDA:
		id = sageMakerPrebuiltECRImageIDByRegion_LDA[region]
	case sageMakerRepositoryXGBoost:
		id = sageMakerPrebuiltECRImageIDByRegion_XGBoost[region]
	case sageMakerRepositoryScikitLearn, sageMakerRepositorySparkML:
		id = sageMakerPrebuiltECRImageIDByRegion_SparkML[region]
	case sageMakerRepositoryTensorFlowServing, sageMakerRepositoryTensorFlowServingEIA:
		id = sageMakerPrebuiltECRImageIDByRegion_TensorFlowServing[region]
	case sageMakerRepositoryMXNetInference,
		sageMakerRepositoryMXNetInferenceEIA,
		sageMakerRepositoryMXNetTraining,
		sageMakerRepositoryPyTorchInference,
		sageMakerRepositoryPyTorchInferenceEIA,
		sageMakerRepositoryPyTorchTraining,
		sageMakerRepositoryTensorFlowInference,
		sageMakerRepositoryTensorFlowInferenceEIA,
		sageMakerRepositoryTensorFlowTraining:
		id = sageMakerPrebuiltECRImageIDByRegion_DeepLearning[region]
	default:
		id = sageMakerPrebuiltECRImageIDByRegion_FactorMachines[region]
	}

	if id == "" {
		return fmt.Errorf("no registry ID available for region (%s) and repository (%s)", region, repo)
	}

	d.SetId(id)
	d.Set("registry_id", id)
	d.Set("registry_path", dataSourceAwsSageMakerPrebuiltECRImageCreatePath(id, region, suffix, repo, d.Get("image_tag").(string)))
	return nil
}

func dataSourceAwsSageMakerPrebuiltECRImageCreatePath(id, region, suffix, repo, imageTag string) string {
	return fmt.Sprintf("%s.dkr.ecr.%s.%s/%s:%s", id, region, suffix, repo, imageTag)
}
