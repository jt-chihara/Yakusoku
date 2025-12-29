#!/bin/bash

echo "Creating S3 bucket for Yakusoku..."
awslocal s3 mb s3://yakusoku
echo "S3 bucket 'yakusoku' created successfully"
