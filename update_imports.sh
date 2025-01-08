#!/bin/bash
find . -name "*.go" -type f -exec sed -i.bak 's|"github.com/username/openipam/|"github.com/lugnut42/openipam/|g' {} +
find . -name "*.go" -type f -exec sed -i.bak 's|"openipam/|"github.com/lugnut42/openipam/|g' {} +
find . -name "*.go.bak" -type f -delete