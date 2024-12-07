# Python implementation
ifndef PYTHON_MK_INCLUDED
PYTHON_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/core/helpers.mk
include $(MAKEFILES_DIR)/core/vars.mk

# Check Python installation
ifeq ($(HAS_PYTHON),)
    $(call log_error,"Python is not installed")
    $(error Python is required)
endif

# Python configuration
PYTHON ?= python3
PIP ?= $(PYTHON) -m pip
VENV ?= .venv
VENV_BIN := $(VENV)/bin
VENV_PYTHON := $(VENV_BIN)/python
VENV_PIP := $(VENV_PYTHON) -m pip

# Python tools configuration
PYTEST_FLAGS ?= -v --tb=short
COVERAGE_FLAGS ?= --cov=src --cov-report=html --cov-report=term-missing
BLACK_FLAGS ?= --line-length=100
ISORT_FLAGS ?= --profile black
MYPY_FLAGS ?= --strict

# Directories
SRC_DIR ?= src
TEST_DIR ?= tests
BUILD_DIR ?= build
DIST_DIR ?= dist

python-info: ## Show Python information
	$(call show_progress,"Python environment information")
	@$(PYTHON) --version
	@$(PIP) --version
	@echo "Virtual environment: $(VENV)"

python-venv: ## Create virtual environment
	$(call show_progress,"Creating virtual environment...")
	@test -d $(VENV) || $(PYTHON) -m venv $(VENV)
	@$(VENV_PIP) install --upgrade pip setuptools wheel

python-deps: python-venv requirements.txt ## Install dependencies
	$(call show_progress,"Installing dependencies...")
	@$(VENV_PIP) install -r requirements.txt

python-dev-deps: python-deps ## Install development dependencies
	$(call show_progress,"Installing development dependencies...")
	@test -f requirements-dev.txt && $(VENV_PIP) install -r requirements-dev.txt || true
	@$(VENV_PIP) install pytest pytest-cov black isort mypy flake8

python-clean: ## Clean build artifacts
	$(call show_progress,"Cleaning Python artifacts...")
	@rm -rf $(BUILD_DIR) $(DIST_DIR) .coverage htmlcov .pytest_cache .mypy_cache
	@find . -type d -name "__pycache__" -exec rm -rf {} +
	@find . -type d -name "*.egg-info" -exec rm -rf {} +
	@find . -type f -name "*.pyc" -delete
	@find . -type f -name "*.pyo" -delete
	@find . -type f -name "*.pyd" -delete

python-clean-venv: ## Remove virtual environment
	$(call show_progress,"Removing virtual environment...")
	@rm -rf $(VENV)

python-test: ## Run tests
	$(call show_test_progress,"Running Python tests...")
	@. $(VENV_BIN)/activate && pytest $(PYTEST_FLAGS) $(TEST_DIR)

python-test-cov: ## Run tests with coverage
	$(call show_test_progress,"Running tests with coverage...")
	@. $(VENV_BIN)/activate && pytest $(PYTEST_FLAGS) $(COVERAGE_FLAGS) $(TEST_DIR)
	@$(OPEN_CMD) htmlcov/index.html

python-lint: ## Run linters
	$(call show_progress,"Running Python linters...")
	@. $(VENV_BIN)/activate && black $(BLACK_FLAGS) --check $(SRC_DIR) $(TEST_DIR)
	@. $(VENV_BIN)/activate && isort $(ISORT_FLAGS) --check-only $(SRC_DIR) $(TEST_DIR)
	@. $(VENV_BIN)/activate && flake8 $(SRC_DIR) $(TEST_DIR)
	@. $(VENV_BIN)/activate && mypy $(MYPY_FLAGS) $(SRC_DIR)

python-format: ## Format Python code
	$(call show_progress,"Formatting Python code...")
	@. $(VENV_BIN)/activate && black $(BLACK_FLAGS) $(SRC_DIR) $(TEST_DIR)
	@. $(VENV_BIN)/activate && isort $(ISORT_FLAGS) $(SRC_DIR) $(TEST_DIR)

python-build: ## Build Python package
	$(call show_build_progress,"Building Python package...")
	@. $(VENV_BIN)/activate && python setup.py build

python-dist: python-build ## Create distribution package
	$(call show_progress,"Creating distribution package...")
	@. $(VENV_BIN)/activate && python setup.py sdist bdist_wheel

python-install: python-build ## Install package locally
	$(call show_progress,"Installing package...")
	@. $(VENV_BIN)/activate && pip install -e .

# Add dependency to project
# Usage: $(call python_add_dep,package_name[,dev])
define python_add_dep
	$(call show_progress,"Adding $(1)...")
	@if [ "$(2)" = "dev" ]; then \
		$(VENV_PIP) install $(1) && \
		$(VENV_PIP) freeze | grep -i "$(1)" >> requirements-dev.txt; \
	else \
		$(VENV_PIP) install $(1) && \
		$(VENV_PIP) freeze | grep -i "$(1)" >> requirements.txt; \
	fi
endef

# Run command in virtual environment
# Usage: $(call python_run,command)
define python_run
	@. $(VENV_BIN)/activate && $(1)
endef

# Standard Python targets that components can use
python-all: python-deps python-lint python-test python-build ## Full Python build cycle

.PHONY: python-info python-venv python-deps python-dev-deps python-clean \
        python-clean-venv python-test python-test-cov python-lint python-format \
        python-build python-dist python-install python-all

endif # PYTHON_MK_INCLUDED
