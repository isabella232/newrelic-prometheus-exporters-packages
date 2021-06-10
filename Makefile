SHELL := /bin/bash
NRI_GENERATOR_PATH="$(PWD)/nri-config-generator"

clean:
	rm -rf dist

build-all:
	@cd exporters; \
	for name in $$(ls -d *) ; do \
		cd $(PWD); \
		make build-$${name}; \
	done

demo:
	source scripts/common_functions.sh; \
	EXPORTER_PATH=exporters/ravendb/exporter.yml; \
	loadVariables; \
	sh exporters/ravendb/build.sh $(PWD);

build-%:
	source scripts/common_functions.sh; \
	EXPORTER_PATH=exporters/$*/exporter.yml; \
	loadVariables; \
	sh scripts/clean.sh $(PWD) $*; \
	sh exporters/$*/build.sh $(PWD); \
	sh scripts/build_generator.sh $(PWD) $*; \
	sh scripts/copy_resources.sh $(PWD) $*

package-%:
	source scripts/common_functions.sh; \
	EXPORTER_PATH=exporters/$*/exporter.yml; \
	loadVariables; \
	sh scripts/package.sh $(PWD) $*

all:
	@cd exporters; \
	for name in $$(ls -d *) ; do \
		cd $(PWD); \
		make	build-$${name}; \
		make 	package-$${name}; \
	done

run:
	sh scripts/run.sh $(PWD)
	docker-compose -f tests/docker-compose.yml up

include $(CURDIR)/nri-config-generator/build/ci.mk