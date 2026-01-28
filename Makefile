.PHONY: help install install-optional install-dev test

help:
	@echo "make install            # core CLI deps"
	@echo "make install-optional   # PNG/PDF/PS, GIF, Qt GUI deps"
	@echo "make install-dev        # test deps"
	@echo "make test               # run tests"

install:
	python3 -m pip install --upgrade pip
	python3 -m pip install -r requirements.txt

install-optional:
	python3 -m pip install -r requirements-optional.txt

install-dev:
	python3 -m pip install -r requirements-dev.txt

test:
	pytest
