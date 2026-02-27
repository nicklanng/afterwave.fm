# Future improvements

Backlog of technical and product improvements we may do later.

## Implementation checklist

- OpenSearch indexing via SQS (decouple write path from search; worker or Lambda consumes queue)
- (Add further items as backlog grows)

---

## OpenSearch indexing via SQS (decouple write path from search)

**Current behavior:** When an artist creates, updates, or deletes a post, the API writes to DynamoDB and then synchronously calls OpenSearch to index the change. If OpenSearch is slow or down, the API request can fail (or hit the server write timeout). If indexing fails after the DynamoDB write, we return an error to the client even though the post is already stored—leading to inconsistent state (post in DB but not in feed) and poor UX.

**Improvement:** Make OpenSearch indexing asynchronous and queue-based.

1. **API:** After a successful DynamoDB write for create/update/delete post, enqueue a message to SQS (e.g. post ID, artist handle, action: index | delete, and any payload needed to index). Return success to the client immediately.
2. **Worker:** A separate process (or Lambda) consumes the queue and calls OpenSearch (same `IndexPost` / delete logic we have today). On success, delete the message; on transient failure, leave it for SQS retries / DLQ.

**Benefits:**

- We never fail the user’s write because of OpenSearch; the DB entry is the source of truth and is never “lost.”
- If OpenSearch is down, messages stay in the queue and are retried until indexing succeeds.
- Feed reads can be eventually consistent (new post appears in “my feed” shortly after the worker runs).

**Implementation sketch:** Add an SQS queue and IAM for the API (send) and worker (receive). In the feed service, after `store.Create` (and equivalent for update/delete), send to SQS instead of calling the indexer directly. New worker binary or Lambda that polls SQS and invokes the existing feed-index OpenSearch calls.
