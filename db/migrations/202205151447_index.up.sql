CREATE INDEX queue_due_date_index ON queue(due_date);

CREATE INDEX queue_is_not_processed_index ON queue(processed_count);