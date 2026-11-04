package handlers

import (
    "log"
    "strings"

    "signal/models"
)

const (
    maxInlineImageBytes      = 32 * 1024 * 1024
    maxTotalInlineImageBytes = 30 * 1024 * 1024
)

func (e *EventHandler) imagesForMessage(groupId string, message models.IncomingMessage) []models.Image {
    if len(message.Attachments) > 0 {
        if images := e.fetchImages(message.Attachments); len(images) > 0 {
            return images
        }
    }

    if message.Quote == nil {
        return nil
    }

    if e.History != nil {
        attachments := e.History.GetAttachments(groupId, message.Quote.ID)
        if images := e.fetchImages(attachments); len(images) > 0 {
            return images
        }
    }

    return e.fetchImages(quotedThumbnails(message.Quote))
}

func (e *EventHandler) fetchImages(attachments []models.Attachment) []models.Image {
    if len(attachments) == 0 {
        return nil
    }
    if e.Attachments == nil {
        log.Printf("Cannot read image attachments because no attachment fetcher is configured")
        return nil
    }

    var images []models.Image
    totalBytes := 0
    for _, attachment := range attachments {
        if !isSupportedImage(attachment) {
            continue
        }
        if attachment.Size > maxInlineImageBytes {
            log.Printf("Skipping image attachment %s because it is too large", attachment.ID)
            continue
        }

        data, err := e.Attachments.GetAttachment(attachment.ID)
        if err != nil {
            log.Printf("failed to fetch attachment %s: %v", attachment.ID, err)
            continue
        }
        if len(data) == 0 || len(data) > maxInlineImageBytes {
            log.Printf("Skipping image attachment %s because its size is invalid", attachment.ID)
            continue
        }
        if totalBytes+len(data) > maxTotalInlineImageBytes {
            log.Printf("Skipping image attachment %s because the total image size is too large", attachment.ID)
            continue
        }
        totalBytes += len(data)

        images = append(images, models.Image{
            ContentType: attachment.ContentType,
            Data:        data,
        })
    }

    return images
}

func quotedThumbnails(quote *models.Quote) []models.Attachment {
    var attachments []models.Attachment
    for _, quoted := range quote.Attachments {
        if quoted.Thumbnail == nil || quoted.Thumbnail.ID == "" {
            continue
        }

        thumbnail := *quoted.Thumbnail
        if thumbnail.ContentType == "" {
            thumbnail.ContentType = quoted.ContentType
        }
        attachments = append(attachments, thumbnail)
    }
    return attachments
}

func isSupportedImage(attachment models.Attachment) bool {
    if attachment.ID == "" {
        return false
    }

    switch strings.ToLower(strings.TrimSpace(attachment.ContentType)) {
    case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp":
        return true
    default:
        return false
    }
}
