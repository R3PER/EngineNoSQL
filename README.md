# EngineNoSQL

## About the Application
EngineNoSQL is a professional application for managing NoSQL databases that enables creating, managing, and manipulating data without the need to define schemas. The application combines simplicity and performance, delivering a complete solution for working with document-format data.

## Screens nad Videos

<img width="1021" height="809" alt="1" src="https://github.com/user-attachments/assets/50ac7e71-f08e-4003-bbc8-b7b8a8731ab0" />
<img width="1022" height="812" alt="2" src="https://github.com/user-attachments/assets/8a09a8dc-cbea-4830-8917-8d44f6a483ab" />
<img width="1022" height="810" alt="3" src="https://github.com/user-attachments/assets/b5df3428-f954-4524-860f-9cf4ca4fad38" />
<img width="1023" height="813" alt="4" src="https://github.com/user-attachments/assets/2e24fab6-1508-4829-9003-70c805aeb20f" />
<img width="1023" height="812" alt="5" src="https://github.com/user-attachments/assets/d3e0bdfb-bd82-4792-a362-351c2d3994cb" />


https://github.com/user-attachments/assets/ac5c633d-0a32-4011-bcd3-9ed57cbcdf1f



## Features

### Database Management
- Creating, deleting, and listing databases
- Creating, deleting, and listing collections
- Document management (adding, updating, deleting)
- Database compaction for performance optimization

### Document Operations
- Advanced queries with filtering, sorting, and pagination
- Field indexing for faster searching
- Counting documents matching criteria

### Data Import/Export
- Importing data from various formats (JSON, CSV)
- Exporting data to external files
- Support for different data formats

### Security
- User authentication system (registration, login)
- Session management
- Database access protection

### Backups
- Creating database backups
- Restoring databases from backups
- Listing available backups

### Statistics and Analysis
- Basic and detailed database statistics
- Performance monitoring

### API Manager
- API access management
- Endpoint configuration

## Architecture and Technologies

### Backend
- **Golang**: Main server-side programming language
- **Wails**: Framework for creating desktop applications using Go and web technologies
- **Custom NoSQL database engine**: Implemented in Go, providing document storage, indexing, and queries

### Frontend
- **React**: JavaScript library for building user interfaces
- **TypeScript**: Typed superset of JavaScript enhancing code safety
- **Ant Design**: UI component library for React
- **Vite**: Frontend application build tool

### Data Structure
- **Documents**: Flexible data structures stored in JSON format
- **Collections**: Grouping of related documents
- **Indexes**: Accelerated searching through field indexing

## Running the Application

### Development Mode
To run the application in development mode, execute the following command in the project directory:
```
wails dev
```

This command will start the Vite server, which provides fast UI refresh. Additionally, a development server will be available at http://localhost:34115, allowing access to Go methods from the browser.

### Building the Application
To build the application in production mode, use the command:

```
wails build
```

This command will create an executable file in the format appropriate for your operating system.

## File Format
The application uses its own `.enosql` file format for storing databases on disk. Databases are serialized and deserialized using JSON format.

## System Requirements
- Supported operating systems: Windows, macOS, Linux
- Minimum hardware requirements: dependent on the size of stored data
