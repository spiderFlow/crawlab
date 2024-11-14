package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type NotificationRequest struct {
	any                            `collection:"notification_requests"`
	BaseModel[NotificationRequest] `bson:",inline"`
	Status                         string               `json:"status" bson:"status"`
	Error                          string               `json:"error,omitempty" bson:"error,omitempty"`
	Title                          string               `json:"title" bson:"title"`
	Content                        string               `json:"content" bson:"content"`
	SenderEmail                    string               `json:"sender_email,omitempty" bson:"sender_email,omitempty"`
	SenderName                     string               `json:"sender_name,omitempty" bson:"sender_name,omitempty"`
	MailTo                         []string             `json:"mail_to,omitempty" bson:"mail_to,omitempty"`
	MailCc                         []string             `json:"mail_cc,omitempty" bson:"mail_cc,omitempty"`
	MailBcc                        []string             `json:"mail_bcc,omitempty" bson:"mail_bcc,omitempty"`
	SettingId                      primitive.ObjectID   `json:"setting_id" bson:"setting_id"`
	ChannelId                      primitive.ObjectID   `json:"channel_id" bson:"channel_id"`
	Setting                        *NotificationSetting `json:"setting,omitempty" bson:"-"`
	Channel                        *NotificationChannel `json:"channel,omitempty" bson:"-"`
	Test                           bool                 `json:"test,omitempty" bson:"test,omitempty"`
}
