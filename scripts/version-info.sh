#!/bin/bash

# version-info.sh - Display current version information

if [ -f VERSION ]; then
    CURRENT_VERSION=$(cat VERSION)
    echo "Current Version: $CURRENT_VERSION"
    
    # Parse the version
    if [[ $CURRENT_VERSION =~ ^([0-9]{4})\.([0-9]{2})\.([0-9]{2})\.([0-9]+)$ ]]; then
        YEAR=${BASH_REMATCH[1]}
        MONTH=${BASH_REMATCH[2]}
        DAY=${BASH_REMATCH[3]}
        PATCH=${BASH_REMATCH[4]}
        
        echo "  Year:  $YEAR"
        echo "  Month: $MONTH"
        echo "  Day:   $DAY"
        echo "  Patch: $PATCH"
        echo ""
        echo "This is patch #$PATCH released on $YEAR-$MONTH-$DAY"
    else
        echo "Version format not recognized as YYYY.MM.DD.PATCH"
    fi
else
    echo "VERSION file not found!"
    exit 1
fi