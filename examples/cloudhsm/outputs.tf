output "hsm_ip_address" {
  value = "${aws_cloudhsm_v2_hsm.cloudhsm_v2_hsm.ip_address}"
}

output "cluster_data_certificate" {
  value = "${data.aws_cloudhsm_v2_cluster.cluster.cluster_certificates.0.cluster_csr}"
}
