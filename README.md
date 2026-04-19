# Event Box API Server
API Server for event management & ticket booking

## API Reference

The Event Box API accepts and returns JSON payloads. Base URL: `https://event-box-huk6.onrender.com`

### System
| Method | Endpoint | Description | Auth Required | Example Usage |
| :--- | :--- | :--- | :---: | :--- |
| `GET` | `/v1/healthcheck` | Returns the system status, environment, and version. | No | `curl -i base_url/v1/healthcheck` |

### Users & Authentication
| Method | Endpoint | Description | Auth Required | Example Payload |
| :--- | :--- | :--- | :---: | :--- |
| `POST` | `/v1/users` | Registers a new user with a role (`organizer` or `customer`). | No | `{"name": "Sohail", "email": "sohail@example.com", "password": "securepassword123", "role": "customer"}` |
| `POST` | `/v1/tokens/authentication` | Authenticates a user and returns a 24-hour Bearer token. | No | `{"email": "sohail@example.com", "password": "securepassword123"}` |

### Events
| Method | Endpoint | Description | Auth Required | Example Payload / Usage |
| :--- | :--- | :--- | :---: | :--- |
| `GET` | `/v1/events` | Lists all events. Supports pagination, sorting, and filtering. | Yes | `GET /v1/events?title=conference&page=1&sort=-date` |
| `POST` | `/v1/events` | Creates a new event. | Yes (Organizer) | `{"title": "Tech Conf", "description": "A great event", "date": "2027-06-15T10:00:00Z", "total_tickets": 200}` |
| `GET` | `/v1/events/:id` | Retrieves the details of a specific event by its ID. | Yes | `curl base_url/v1/events/1` |
| `PATCH` | `/v1/events/:id` | Partially updates an event. Only the organizer who created it may update. | Yes (Organizer) | `{"title": "Updated Title", "description": "New description"}` |

### Bookings
| Method | Endpoint | Description | Auth Required | Example Payload / Usage |
| :--- | :--- | :--- | :---: | :--- |
| `POST` | `/v1/bookings` | Books a ticket for an event. Atomic — checks availability and decrements in one transaction. | Yes (Customer) | `{"event_id": 1}` |
| `GET` | `/v1/bookings` | Lists bookings. Customers see their own bookings. Organizers see all bookings for their events. | Yes | `curl base_url/v1/bookings` |

### Metrics
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `GET` | `/debug/vars` | Expvar debug metrics (goroutines, DB stats, version). | No |

> **Note on Authentication:** For routes requiring authentication, include the token from `/v1/tokens/authentication` in your request header:
> `Authorization: Bearer <your-authentication-token>`

> **Note on Roles:** Users register with a `role` of either `organizer` or `customer`. Organizers can create and manage events. Customers can book tickets.

## Running Locally

```bash
# Clone the repository
git clone https://github.com/thecodephilic-guy/event-box.git
cd event-box

# Set up environment variables (create a .env file)
echo 'DATABASE_URL=postgres://user:pass@localhost/eventbox?sslmode=disable' > .env

# Run database migrations
migrate -path=./migrations -database=$DATABASE_URL up

# Start the server
go run ./cmd/api
```

## Running Tests

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -v -cover
```
