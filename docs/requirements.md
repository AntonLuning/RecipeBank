## Requirements

### Non-Functional
- **Authentication**
    - Implement stateless authentication for "non-viewing" functionality
    - Utilize pre-shared credentials
    - Implement account lockout after X consecutive failed login attempts
- **User Interface**
    - Support responsive design for three screen sizes: mobile, tablet, and desktop
- **Logging and Monitoring**
    - Log all user interactions
    - Provide usage and performance metrics via Prometheus
- **Performance and Security**
    - Implement rate limiting to prevent system overload
- **Deployment**
    - Package application as a Docker container
    - Configure via environment variables

### Functional
- **UI - Home Screen**
    - Display recipes as a grid of cards with filtering options
    - Include a prominent "Add New Recipe" button
    - *Optional: Implement search functionality and recipe sharing feature*
- **UI - Recipe Detail Page**
    - Access via clicking a recipe card or the "Add New Recipe" button
    - Display full recipe information
    - Include an edit button (auto-selected if "Add New Recipe")
    - Include a delete button when editing
    - Allow adjustment of ingredient quantities based on serving size
- **UI - User log in**
    - If the user wants to add or edit a recipe (and not already logged in) - this page will be used
    Require user authentication to access
- **Recipe Layout**
    - Utilize pre-defined templates, selectable during recipe edit
- **API Architecture**
    - Separate business logic API from UI API
- **User Management**    
    - Manage users manually through backend database

### External Interface
- **Database**
    - Utilize MongoDB instance
    - With automatic, preferably off-site backups
- **Monitoring and Visualization**
    - Integrate with Prometheus for metric collection
    - Provide Grafana dashboards for metric visualization
- **Security**
    - Utilize reverse-proxy with TLS handling
- **Documentation**
    - Generate automatic API documentation
    