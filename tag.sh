#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Functions
show_help() {
    echo "Usage: $0 [create|delete] <tag>"
    echo ""
    echo "Commands:"
    echo "  create <tag>    Create and push a new tag (e.g., v1.0.0)"
    echo "  delete <tag>    Delete tag locally and remotely"
    echo ""
    echo "Examples:"
    echo "  $0 create v1.0.0"
    echo "  $0 delete v1.0.0"
}

create_tag() {
    local tag=$1

    if [ -z "$tag" ]; then
        echo -e "${RED}Error: Tag name is required${NC}"
        show_help
        exit 1
    fi

    # Check if tag already exists locally
    if git rev-parse "$tag" >/dev/null 2>&1; then
        echo -e "${RED}Error: Tag '$tag' already exists locally${NC}"
        exit 1
    fi

    echo -e "${GREEN}Creating tag: $tag${NC}"
    git tag "$tag"

    echo -e "${GREEN}Pushing tag to remote...${NC}"
    git push origin "$tag"

    echo -e "${GREEN}✅ Tag '$tag' created and pushed successfully!${NC}"
    echo -e "${YELLOW}GitHub Actions will now build and create the release.${NC}"
}

delete_tag() {
    local tag=$1

    if [ -z "$tag" ]; then
        echo -e "${RED}Error: Tag name is required${NC}"
        show_help
        exit 1
    fi

    # Confirm deletion
    read -p "Are you sure you want to delete tag '$tag'? (y/N): " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Deletion cancelled${NC}"
        exit 0
    fi

    # Delete local tag if exists
    if git rev-parse "$tag" >/dev/null 2>&1; then
        echo -e "${GREEN}Deleting local tag: $tag${NC}"
        git tag -d "$tag"
    else
        echo -e "${YELLOW}Tag '$tag' does not exist locally${NC}"
    fi

    # Delete remote tag if exists
    if git ls-remote --tags origin | grep -q "refs/tags/$tag"; then
        echo -e "${GREEN}Deleting remote tag: $tag${NC}"
        git push origin ":refs/tags/$tag"
    else
        echo -e "${YELLOW}Tag '$tag' does not exist on remote${NC}"
    fi

    echo -e "${GREEN}✅ Tag '$tag' deleted successfully!${NC}"
}

list_tags() {
    echo -e "${GREEN}Recent tags:${NC}"
    git tag -l --sort=-version:refname | head -10
}

# Main
case "${1:-}" in
    create)
        create_tag "$2"
        ;;
    delete)
        delete_tag "$2"
        ;;
    list)
        list_tags
        ;;
    -h|--help|help)
        show_help
        ;;
    *)
        echo -e "${RED}Error: Invalid command${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac
