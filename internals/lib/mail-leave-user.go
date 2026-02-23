package lib

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"time"

	"github.com/Abhishekh669/backend/internals/models"
)

type EmailConfig struct {
	SMTPHost       string
	SMTPPort       string
	SenderEmail    string
	SenderPassword string
}

type LeaveEmailData struct {
	EmployeeName      string
	EmployeeEmail     string
	Status            models.LeaveStatus
	StartDate         time.Time
	EndDate           time.Time
	TotalDays         int
	Message           string
	SupervisorMessage *string
	Year              int
}

type MailService struct {
	config EmailConfig
}

func NewMailService(config EmailConfig) *MailService {
	return &MailService{
		config: config,
	}
}

// SendLeaveStatusEmail sends an email about leave request status
func (m *MailService) SendLeaveStatusEmail(to string, data LeaveEmailData) error {
	subject := fmt.Sprintf("Leave Request %s", data.Status)

	// Get email template
	emailBody, err := m.getLeaveEmailTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to generate email template: %w", err)
	}

	// Compose full email
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"%s\r\n"+
		"%s", to, subject, mime, emailBody))

	// SMTP authentication
	auth := smtp.PlainAuth("", m.config.SenderEmail, m.config.SenderPassword, m.config.SMTPHost)

	// Send email
	err = smtp.SendMail(
		m.config.SMTPHost+":"+m.config.SMTPPort,
		auth,
		m.config.SenderEmail,
		[]string{to},
		msg,
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (m *MailService) getLeaveEmailTemplate(data LeaveEmailData) (string, error) {
	// Format dates
	startDate := data.StartDate.Format("Monday, January 2, 2006")
	endDate := data.EndDate.Format("Monday, January 2, 2006")

	// Calculate total days automatically from start and end date
	totalDays := int(data.EndDate.Sub(data.StartDate).Hours()/24) + 1

	// Get year from start date
	year := data.StartDate.Year()

	// Determine color and status string based on status
	var statusColor string
	var statusString string

	switch data.Status {
	case models.LeaveApproved:
		statusColor = "#4CAF50" // Green
		statusString = "Accepted"
	case models.LeaveRejected:
		statusColor = "#F44336" // Red
		statusString = "Rejected"

	default:
		statusColor = "#6C757D" // Gray for unknown
		statusString = string(data.Status)
	}

	// HTML template
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background: #ffffff;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            border-bottom: 3px solid {{.StatusColor}};
        }
        .header h1 {
            margin: 0;
            color: {{.StatusColor}};
            font-size: 24px;
        }
        .content {
            padding: 30px;
        }
        .status-badge {
            display: inline-block;
            padding: 8px 20px;
            background: {{.StatusColor}};
            color: white;
            border-radius: 25px;
            font-weight: bold;
            text-transform: uppercase;
            font-size: 14px;
            margin-bottom: 20px;
        }
        .info-box {
            background: #f8f9fa;
            border-left: 4px solid {{.StatusColor}};
            padding: 15px;
            margin: 20px 0;
            border-radius: 0 5px 5px 0;
        }
        .info-row {
            margin-bottom: 10px;
            padding-bottom: 10px;
            border-bottom: 1px solid #dee2e6;
        }
        .info-row:last-child {
            border-bottom: none;
            margin-bottom: 0;
            padding-bottom: 0;
        }
        .info-label {
            font-weight: bold;
            color: #666;
            display: inline-block;
            width: 120px;
        }
        .info-value {
            color: #333;
        }
        .message-box {
            background: #fff;
            border: 1px solid #dee2e6;
            padding: 15px;
            margin: 20px 0;
            border-radius: 5px;
        }
        .supervisor-message {
            background: #fff3cd;
            border: 1px solid #ffeeba;
            color: #856404;
            padding: 15px;
            margin: 20px 0;
            border-radius: 5px;
        }
        .footer {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            font-size: 12px;
            color: #666;
            border-top: 1px solid #dee2e6;
        }
        .date-range {
            font-size: 16px;
            font-weight: bold;
            color: {{.StatusColor}};
        }
        @media only screen and (max-width: 600px) {
            .container {
                margin: 10px;
            }
            .content {
                padding: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Leave Request {{.StatusString}}</h1>
        </div>
        
        <div class="content">
            <div style="text-align: center;">
                <span class="status-badge">{{.StatusString}}</span>
            </div>
            
            <p>Dear <strong>{{.EmployeeName}}</strong>,</p>
            
            <p>Your leave request has been <strong style="color: {{.StatusColor}};">{{.StatusString}}</strong>.</p>
            
            <div class="info-box">
                <div class="info-row">
                    <span class="info-label">Employee:</span>
                    <span class="info-value">{{.EmployeeName}} ({{.EmployeeEmail}})</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Leave Period:</span>
                    <span class="info-value date-range">
                        {{.StartDate}} to {{.EndDate}}
                    </span>
                </div>
                <div class="info-row">
                    <span class="info-label">Total Days:</span>
                    <span class="info-value">{{.TotalDays}} day(s)</span>
                </div>
            </div>
            
            {{if .Message}}
            <div class="message-box">
                <strong>Your Message:</strong>
                <p style="margin: 10px 0 0 0;">{{.Message}}</p>
            </div>
            {{end}}
            
            {{if .SupervisorMessage}}
            <div class="supervisor-message">
                <strong>Supervisor's Feedback:</strong>
                <p style="margin: 10px 0 0 0;">{{.SupervisorMessage}}</p>
            </div>
            {{end}}
            
            <p style="margin-top: 30px;">
                {{if eq .StatusString "Accepted"}}
                    Your leave has been approved. Have a great time off!
                {{else if eq .StatusString "Rejected"}}
                    Your leave request has been declined. Please contact your supervisor for more information.
                {{else if eq .StatusString "Pending"}}
                    Your leave request is currently being reviewed. We'll notify you once a decision is made.
                {{else}}
                    Your leave request status has been updated.
                {{end}}
            </p>
        </div>
        
        <div class="footer">
            <p>This is an automated message from the Attendance Management System.</p>
            <p>&copy; {{.Year}} Your Company Name. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	// Parse template
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	// Template data with automatically calculated values
	templateData := struct {
		LeaveEmailData
		StatusColor  string
		StatusString string
		StartDate    string
		EndDate      string
		TotalDays    int
		Year         int
	}{
		LeaveEmailData: data,
		StatusColor:    statusColor,
		StatusString:   statusString,
		StartDate:      startDate,
		EndDate:        endDate,
		TotalDays:      totalDays,
		Year:           year,
	}

	var body bytes.Buffer
	if err := t.Execute(&body, templateData); err != nil {
		return "", err
	}

	return body.String(), nil
}
