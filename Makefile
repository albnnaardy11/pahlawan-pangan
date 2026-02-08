# Makefile for Go Code Quality Automation

.PHONY: all fmt lint vet fix license

# Default target
all: fmt vet lint

# 1. Format Code (gofmt -s -w)
fmt:
	@echo "🎨 Running gofmt..."
	gofmt -s -w .

# 2. Go Vet (Logical Errors)
vet:
	@echo "🔍 Running go vet..."
	go vet ./...

# 3. Linting (requires golangci-lint installed)
lint:
	@echo "🧹 Running golangci-lint..."
	golangci-lint run ./...

# 4. Auto-Fix (Go Imports & Lint adjustments)
fix:
	@echo "🔧 Running auto-fixers..."
	goimports -w .
	golangci-lint run --fix ./...

# 5. Add Missing Comments (Placeholder - requires 'golint' or custom script)
# Ideally use a tool like 'gomodifytags' or similar for structural changes,
# but for docs, manual review or AI-assisted batching is often better.
# Here we just list exported functions missing comments.
audit-docs:
	@echo "📝 Auditing undocumented exported symbols..."
	@golint ./... | grep "exported" || true

# 6. Generate License (MIT)
license:
	@echo "📜 Generating MIT License..."
	@echo "MIT License" > LICENSE
	@echo "" >> LICENSE
	@echo "Copyright (c) $(shell date +%Y) Pahlawan Pangan Contributors" >> LICENSE
	@echo "" >> LICENSE
	@echo "Permission is hereby granted, free of charge, to any person obtaining a copy" >> LICENSE
	@echo "of this software and associated documentation files (the \"Software\"), to deal" >> LICENSE
	@echo "in the Software without restriction, including without limitation the rights" >> LICENSE
	@echo "to use, copy, modify, merge, publish, distribute, sublicense, and/or sell" >> LICENSE
	@echo "copies of the Software, and to permit persons to whom the Software is" >> LICENSE
	@echo "furnished to do so, subject to the following conditions:" >> LICENSE
	@echo "" >> LICENSE
	@echo "The above copyright notice and this permission notice shall be included in all" >> LICENSE
	@echo "copies or substantial portions of the Software." >> LICENSE
	@echo "" >> LICENSE
	@echo "THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR" >> LICENSE
	@echo "IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY," >> LICENSE
	@echo "FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE" >> LICENSE
	@echo "AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER" >> LICENSE
	@echo "LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM," >> LICENSE
	@echo "OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE" >> LICENSE
	@echo "SOFTWARE." >> LICENSE
