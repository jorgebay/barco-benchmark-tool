variable "cluster_length" {
  description = "Number of brokers in the cluster"
  default = 3
  type = number
}

variable "cluster_instance_type" {
  description = "The instance type for the brokers"
  default = "t4g.micro"
}

variable "cluster_arch" {
  default = "arm64"
}

variable "client_length" {
  description = "Number of clients"
  default = 1
  type = number
}

variable "client_instance_type" {
  default = "c6in.8xlarge"
}

variable "key_name" {
  default = "my_key"
}
