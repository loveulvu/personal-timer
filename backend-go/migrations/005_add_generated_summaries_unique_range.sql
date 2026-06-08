ALTER TABLE generated_summaries
ADD UNIQUE KEY uq_generated_summaries_type_range (summary_type, start_date, end_date);
