#!/bin/bash

# Source the functions from pre-commit hook
source .git/hooks/pre-commit

# Test the lint_yaml_files function
lint_yaml_files
