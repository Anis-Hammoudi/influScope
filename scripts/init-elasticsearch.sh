#!/bin/bash
# Initialize Elasticsearch index for InfluScope

ELASTIC_URL="${ELASTIC_URL:-http://localhost:9200}"
INDEX_NAME="influencers"

echo "Creating Elasticsearch index: $INDEX_NAME"

curl -X PUT "$ELASTIC_URL/$INDEX_NAME" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "analyzer": {
        "username_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "username": {
        "type": "text",
        "analyzer": "username_analyzer",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "platform": { "type": "keyword" },
      "display_name": { "type": "text" },
      "bio": { "type": "text" },
      "followers_count": { "type": "long" },
      "following_count": { "type": "long" },
      "posts_count": { "type": "long" },
      "engagement_rate": { "type": "float" },
      "verified": { "type": "boolean" },
      "categories": { "type": "keyword" },
      "created_at": { "type": "date" },
      "updated_at": { "type": "date" }
    }
  }
}
'

echo ""
echo "Elasticsearch index created successfully!"
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo

# Environment files
.env
.env.local
.env.*.local

# Logs
*.log
logs/

# OS files
.DS_Store
Thumbs.db

# Vendor (if you commit deps, remove this)
vendor/

# Build output
bin/
dist/

# Docker
.docker/

# Temporary files
tmp/
temp/

