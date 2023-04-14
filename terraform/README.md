# Terraform PolarStreams Cluster for Benchmarking

Set your [AWS key pair name](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/create-key-pairs.html).

```shell
export TF_VAR_key_name=YOUR_KEY_NAME_HERE
```

Set the instances types and apply.

```shell
export TF_VAR_cluster_instance_type=c6i.xlarge
export TF_VAR_cluster_arch=amd64

terraform apply
```

## Access the instances

Broker instances.

```shell
ssh -i ~/Downloads/your_key.pem -o StrictHostKeyChecking=no ubuntu@$(terraform output -json brokers_ips | jq -r '.[0]')
ssh -i ~/Downloads/your_key.pem -o StrictHostKeyChecking=no ubuntu@$(terraform output -json brokers_ips | jq -r '.[1]')
ssh -i ~/Downloads/your_key.pem -o StrictHostKeyChecking=no ubuntu@$(terraform output -json brokers_ips | jq -r '.[2]')
```

Client instances.

```shell
ssh -i ~/Downloads/your_key.pem ubuntu@$(terraform output -json clients_ips | jq -r '.[0]')
```
