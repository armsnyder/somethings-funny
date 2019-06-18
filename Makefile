.PHONY: sds
sds:
	./script/release_sds.sh

.PHONY: mic
mic:
	./script/deploy_mic.sh

.PHONY: dns
dns:
	./script/flush_dns.sh

.PHONY: envoy
envoy:
	./script/release_envoy.sh

.PHONY: service
service:
	./script/release_service.sh

.PHONY: restart
restart:
	./script/restart_service.sh

.PHONY: tf
tf:
	./script/terraform_apply.sh

.PHONY: sampler
sampler:
	./script/deploy_sampler.sh
