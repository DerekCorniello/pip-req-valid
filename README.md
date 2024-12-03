![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Vue.js](https://img.shields.io/badge/vue.js-%234FC08D.svg?style=for-the-badge&logo=vue.js&logoColor=white)
![Vite](https://img.shields.io/badge/vite-%23646CFF.svg?style=for-the-badge&logo=vite&logoColor=white)
![CSS3](https://img.shields.io/badge/css3-%231572B6.svg?style=for-the-badge&logo=css3&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)
![AWS](https://img.shields.io/badge/aws-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)

# ReqInspect

ReqInspect is an open-source application designed to simplify and streamline request inspection and debugging. It provides a full-stack solution with a **Vue.js** frontend and a **Go** backend, offering a responsive and powerful interface for inspecting HTTP requests and responses in real time.

## Infrastructure Overview

The ReqInspect application is designed to be robust, scalable, and secure. Here’s an outline of the infrastructure setup for deployment:

### Frontend
- **Framework**: Built with Vue.js for a modern, dynamic, and responsive user interface.
- **Hosting**: Deployed on **AWS Amplify**, taking advantage of its managed hosting services for SPAs (Single Page Applications).
  - CDN-backed hosting for fast and reliable global delivery.
  - HTTPS enabled by default for secure communication.
  - Automatic builds and deployments on commits to the main branch.
  - Automatic Domain Resolution linked with **Route53**

### Backend
- **Framework**: Powered by a Go API that efficiently handles HTTP requests.
- **Hosting**: Deployed on an **AWS EC2 instance** to ensure flexibility and scalability.
  - Rate Limited API to mitigate attacks and spamming of the backend
  - Token-based authentication implemented for enhanced security:
    - Clients must obtain an authentication token before sending requests.
    - Tokens ensure secure file uploads and authenticated API interactions.

### Networking & Security
- **Domain**: The application is accessible through [reqinspect.com](https://reqinspect.com) and [www.reqinspect.com](https://www.reqinspect.com).
- **Load Balancing**:
  - An **Application Load Balancer (ALB)** distributes incoming requests to the backend instances.
  - HTTPS termination occurs at the ALB to ensure end-to-end encryption.
- **Authentication & Authorization**:
  - The backend enforces token-based authentication for all API endpoints.
  - Expired or invalid tokens are rejected to maintain API security.
- **Rate Limiting**:
  - The backend limits the amount of pings per hour in order to mitigate attacks

### Monitoring & Logging
- **Logging**: Both the frontend and backend are integrated with AWS CloudWatch for centralized log collection and analysis.
- **Metrics**: CloudWatch monitors system performance, traffic patterns, and error rates.
- **Alerting**: Alarms notify the administrator of unusual activities or potential system failures.

### Scalability
- **Frontend**: AWS Amplify ensures automatic scaling for traffic surges without manual intervention.
- **Backend**: The EC2 instance is provisioned with auto-scaling rules, ensuring optimal performance even under heavy load.

## Contributing
We welcome contributions! If you'd like to suggest a feature, report a bug, or make a pull request, please follow the contribution guidelines in this repository.

---

#### License:
Licensed under Apache 2.0 License

## Connect with Me!
[![LinkedIn](https://img.shields.io/badge/LinkedIn-%230A66C2.svg?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/derek-corniello)
[![GitHub](https://img.shields.io/badge/GitHub-%23121011.svg?style=for-the-badge&logo=github&logoColor=white)](https://github.com/derekcorniello)
[![X](https://img.shields.io/badge/X-%231DA1F2.svg?style=for-the-badge&logo=x&logoColor=white)](https://x.com/derekcorniello)
