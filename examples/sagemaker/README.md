# AWS SageMaker Example

This example takes the [example model provided by AWS](https://github.com/awslabs/amazon-sagemaker-examples/blob/master/advanced_functionality/scikit_bring_your_own/scikit_bring_your_own.ipynb)
to show how to deploy your own machine learning algorithm into a SageMaker container using Terraform.


### Wrap model in Docker container and upload to [ECR](https://aws.amazon.com/ecr/)

Get the SageMaker example model from AWS:

    git clone https://github.com/awslabs/amazon-sagemaker-examples.git
    cd amazon-sagemaker-examples/advanced_functionality/scikit_bring_your_own/container/

Export credentials for your account:

    export AWS_ACCESS_KEY_ID=<your-access-key-id>
    export AWS_SECRET_ACCESS_KEY=<your-secret-access-key>
    
Create docker container and push it to ECR:

    ./build_and_push.sh foo

### Deploy model and run test prediction call

In the directory where this README is located, run the following:

    terraform init
    terraform apply
   

Go back to `amazon-sagemaker-examples/advanced_functionality/scikit_bring_your_own/container/` and make a test call to the deployed model:

    aws runtime.sagemaker invoke-endpoint --endpoint-name terraform-sagemaker-example \
        --body "`cat ./local_test/payload.csv`" --content-type "text/csv" "output.dat"

Have a look the predicted values:

    cat output.dat

Destroy the deployed model:

    terraform destroy
