#!/usr/bin/env python3
"""
Validator for CodeRabbit YAML configuration files.
Checks for common schema issues based on CodeRabbit's expected format.
"""

import yaml
import sys
import os

def validate_coderabbit_yaml(filepath):
    """Validate CodeRabbit YAML configuration against expected schema."""
    errors = []
    
    try:
        with open(filepath, 'r') as f:
            config = yaml.safe_load(f)
    except yaml.YAMLError as e:
        return [f"YAML syntax error: {e}"]
    except FileNotFoundError:
        return [f"File not found: {filepath}"]
    
    # Check if reviews section exists
    if 'reviews' not in config:
        errors.append("Missing 'reviews' section")
        return errors
    
    reviews = config['reviews']
    
    # Validate labeling_instructions if present
    if 'labeling_instructions' in reviews:
        instructions = reviews['labeling_instructions']
        
        if not isinstance(instructions, list):
            errors.append("'reviews.labeling_instructions' must be a list")
        else:
            for i, instruction in enumerate(instructions):
                # Check if instruction is an object (dict) instead of string
                if isinstance(instruction, str):
                    errors.append(
                        f"Expected object, received string at \"reviews.labeling_instructions[{i}]\""
                    )
                elif isinstance(instruction, dict):
                    # Validate object structure
                    if 'label' not in instruction:
                        errors.append(
                            f"Missing 'label' field at \"reviews.labeling_instructions[{i}]\""
                        )
                    elif not isinstance(instruction['label'], str):
                        errors.append(
                            f"'label' must be a string at \"reviews.labeling_instructions[{i}]\""
                        )
                    
                    if 'instructions' not in instruction:
                        errors.append(
                            f"Missing 'instructions' field at \"reviews.labeling_instructions[{i}]\""
                        )
                    elif not isinstance(instruction['instructions'], str):
                        errors.append(
                            f"'instructions' must be a string at \"reviews.labeling_instructions[{i}]\""
                        )
                    elif len(instruction['instructions']) > 3000:
                        errors.append(
                            f"'instructions' exceeds 3000 character limit at \"reviews.labeling_instructions[{i}]\""
                        )
                else:
                    errors.append(
                        f"Invalid type at \"reviews.labeling_instructions[{i}]\": "
                        f"expected object, got {type(instruction).__name__}"
                    )
    
    # Additional validations can be added here
    # For example, checking auto_review structure, etc.
    
    return errors

def main():
    """Main entry point."""
    if len(sys.argv) < 2:
        filepath = '.coderabbit.yaml'
    else:
        filepath = sys.argv[1]
    
    # Get absolute path
    filepath = os.path.abspath(filepath)
    
    print(f"Validating CodeRabbit configuration: {filepath}")
    
    errors = validate_coderabbit_yaml(filepath)
    
    if errors:
        print("\n❌ Validation failed with the following errors:")
        for error in errors:
            print(f"  - {error}")
        sys.exit(1)
    else:
        print("✅ CodeRabbit YAML configuration is valid")
        sys.exit(0)

if __name__ == "__main__":
    main()