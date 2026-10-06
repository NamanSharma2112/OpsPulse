-- M5: detected anomalies, and the incidents they raised.
BEGIN;

CREATE TABLE anomalies (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    metric          text NOT NULL,
    -- The bucket that was judged, not when the judgement ran. Re-evaluating
    -- the same bucket must not raise the same anomaly twice.
    bucket_start    timestamptz NOT NULL,
    observed        double precision NOT NULL,
    baseline_median double precision NOT NULL,
    baseline_spread double precision NOT NULL,
    deviation       double precision NOT NULL,
    direction       text NOT NULL CHECK (direction IN ('above', 'below')),
    -- Written when the anomaly was detected, so the explanation reflects the
    -- numbers at that moment rather than today's baseline.
    explanation     text NOT NULL,
    incident_id     uuid REFERENCES incidents (id) ON DELETE SET NULL,
    detected_at     timestamptz NOT NULL DEFAULT now(),

    UNIQUE (project_id, metric, bucket_start)
);

CREATE INDEX anomalies_project_detected_idx ON anomalies (project_id, detected_at DESC);

COMMIT;
