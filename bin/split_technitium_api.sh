#!/bin/bash

# Create directory for split files if it doesn't exist
mkdir -p .ai/docs/technitium-api

# Create TOC file
TOC_FILE=".ai/docs/technitium-api/00_table_of_contents.md"
echo "# Technitium DNS Server API Documentation - Table of Contents" > $TOC_FILE
echo "" >> $TOC_FILE
echo "This document contains links to all sections of the Technitium DNS Server API documentation." >> $TOC_FILE
echo "" >> $TOC_FILE

# Process the original file
INPUT_FILE=".ai/docs/technitium-api.md"

# Read the file and split by H2 sections
section_num=1
section_content=""
section_title=""
in_section=false

# First pass to get all section titles for TOC
while IFS= read -r line; do
    if [[ $line =~ ^"## " ]]; then
        section_title="${line#"## "}"
        # Format for TOC entry
        padded_num=$(printf "%02d" $section_num)
        echo "- [$section_title](./${padded_num}_${section_title// /_}.md)" >> $TOC_FILE
        ((section_num++))
    fi
done < "$INPUT_FILE"

# Reset for second pass
section_num=1
section_content=""
section_title=""
in_section=false

# Second pass to create individual files
while IFS= read -r line; do
    if [[ $line =~ ^"## " ]]; then
        # Save previous section if it exists
        if $in_section; then
            padded_num=$(printf "%02d" $((section_num-1)))
            filename=".ai/docs/technitium-api/${padded_num}_${section_title// /_}.md"
            echo "$section_content" > "$filename"
        fi
        
        # Start new section
        section_title="${line#"## "}"
        section_content="# Technitium DNS Server API - ${section_title}

${line}
"
        in_section=true
        ((section_num++))
    elif $in_section; then
        section_content="${section_content}${line}
"
    fi
done < "$INPUT_FILE"

# Save the last section
if $in_section; then
    padded_num=$(printf "%02d" $((section_num-1)))
    filename=".ai/docs/technitium-api/${padded_num}_${section_title// /_}.md"
    echo "$section_content" > "$filename"
fi

echo "Split completed. $((section_num-1)) files created in .ai/docs/technitium-api/"