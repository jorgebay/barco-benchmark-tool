provider "aws" {
  region = "us-east-2"
}

resource aws_vpc r1 {
  cidr_block = "10.0.0.0/16"
  enable_dns_hostnames = true
  tags = {
    Name = "Benchmarks - VPC"
  }
}

data aws_availability_zones r1 {}

resource aws_subnet r1_az1 {
  vpc_id = aws_vpc.r1.id
  cidr_block = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.r1.names[0]
}

resource aws_security_group sg_default_r1 {
  name        = "sg_benchmarks_default_r1"
  vpc_id      = aws_vpc.r1.id

  # outbound internet access
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # SSH from the internet
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Ingress from the peers
  ingress {
    from_port   = 9250
    to_port     = 9300
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/24"]
  }
}

resource aws_internet_gateway agw_r1 {
  vpc_id = aws_vpc.r1.id
  tags = {
    Name = "Benchmarks - IG R1",
  }
}

resource aws_route internet_access_r1 {
  route_table_id         = aws_vpc.r1.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.agw_r1.id
}

resource aws_instance clients {
  count = var.client_length
  ami = data.aws_ami.ubuntu_amd64.id
  instance_type = var.client_instance_type
  user_data = file("client_install.sh")
  subnet_id = aws_subnet.r1_az1.id
  vpc_security_group_ids = [aws_security_group.sg_default_r1.id]
  associate_public_ip_address = true
  key_name = var.key_name
  tags = {
    Name        = "Benchmarks - Client${count.index}",
    Terraform   = "true"
  }
}

output "clients_ips" {
  value = aws_instance.clients[*].public_ip
}

resource aws_instance brokers {
  count = var.cluster_length
  instance_type = var.cluster_instance_type
  ami = local.amis[var.cluster_arch]
  user_data = file("broker_install_${var.cluster_arch}.sh")
  subnet_id = aws_subnet.r1_az1.id
  vpc_security_group_ids = [aws_security_group.sg_default_r1.id]
  associate_public_ip_address = true
  key_name = var.key_name
  private_ip = "10.0.0.10${count.index}"
  credit_specification {
    cpu_credits = "standard" # or "unlimited"
  }
  tags = {
    Name        = "Benchmarks - Broker${count.index}",
    Terraform   = "true"
  }
}

output "brokers_ips" {
  value = [
    for b in aws_instance.brokers:
      b.public_ip
  ]
}

resource "aws_volume_attachment" "broker_va" {
  count = var.cluster_length
  device_name = "/dev/sdh"
  volume_id   = aws_ebs_volume.broker_volume[count.index].id
  instance_id = aws_instance.brokers[count.index].id
}

resource "aws_ebs_volume" "broker_volume" {
  count = var.cluster_length
  availability_zone = data.aws_availability_zones.r1.names[0]
  size              = 100
  type              = "gp3"
  iops              = 3000
}
