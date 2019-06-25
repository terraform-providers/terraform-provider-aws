output "elb_dns_name" {
  value = "${aws_elb.web.dns_name}"
}

output "web_public_ip" {
  value = "${aws_instance.web.public_ip}"
}

output "web_private_ip" {
  value = "${aws_instance.web.private_ip}"
}

output "db_public_ip" {
  value = "${aws_instance.db.public_ip}"
}

output "db_private_ip" {
  value = "${aws_instance.db.private_ip}"
}

