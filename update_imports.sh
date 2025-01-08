#!/bin/bash
find . -name "*.go" -type f -exec sed -i.bak 's|"openipam/internal/|"github.com/username/openipam/internal/|g' {} +
find . -name "*.go.bak" -type f -delete