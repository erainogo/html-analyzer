# GitLab CI/CD Pipeline for HTML Analyzer

This guide outlines how to set up a continuous integration and continuous deployment (CI/CD) pipeline for the HTML Analyzer project using GitLab CI/CD.

## Pipeline Overview

The GitLab CI/CD pipeline will:

1. Build and deps & test & lint the application
2. Publish to registry
3. Deploy to testing
4. Deploy to staging
5. Deploy to production (with manual approval)

## 🚀 Deployment Strategies

### Blue-Green Deployment

For blue-green deployments with AWS ECS:

1. Create two ECS service tasks (blue and green)
2. Deploy to the inactive environment
3. Run tests against the new environment
4. Switch traffic to the new environment using a load balancer
5. Terminate the old environment if successful