"""Tests for the vLLM client and reasoning stripping.

Run with:  python -m pytest test_vllm_client.py -v
     or:   python -m unittest test_vllm_client -v
"""

import sys
import unittest
from unittest import mock

# torch / numpy / huggingface_hub are optional at runtime and absent from the
# production venv; make the import path identical here so the tests exercise the
# same code path as production.
for _optional in ("torch", "numpy", "huggingface_hub"):
    sys.modules.setdefault(_optional, None)

from hrm_sapient import VLLMClient, strip_reasoning  # noqa: E402


def fake_response(status=200, payload=None, text=""):
    """Build a stand-in for a requests.Response."""
    resp = mock.Mock()
    resp.status_code = status
    resp.json.return_value = payload or {}
    resp.text = text
    return resp


def chat_payload(content, reasoning=None):
    message = {"content": content}
    if reasoning is not None:
        message["reasoning"] = reasoning
    return {"choices": [{"message": message}]}


class StripReasoningTest(unittest.TestCase):
    def test_removes_closed_think_block(self):
        self.assertEqual(strip_reasoning("<think>bla</think>Paris"), "Paris")

    def test_removes_unclosed_block_from_truncated_output(self):
        # max_tokens cut the answer mid-thought: everything after <think> is
        # reasoning, never an answer.
        self.assertEqual(strip_reasoning("\n\n<think>je commence a red"), "")

    def test_keeps_plain_text_untouched(self):
        self.assertEqual(strip_reasoning("une reponse directe"), "une reponse directe")

    def test_handles_empty_input(self):
        self.assertEqual(strip_reasoning(""), "")

    def test_preserves_json_after_reasoning(self):
        self.assertEqual(strip_reasoning('<think>x</think>{"k": 1}'), '{"k": 1}')


class VLLMClientTest(unittest.TestCase):
    def setUp(self):
        self.client = VLLMClient("http://serveur/v1", "un-modele")

    def test_uses_chat_completions_endpoint(self):
        with mock.patch("hrm_sapient.requests.post",
                        return_value=fake_response(payload=chat_payload("ok"))) as post:
            self.client.generate("prompt")

        url = post.call_args[0][0]
        self.assertTrue(url.endswith("/chat/completions"),
                        f"appel sur {url}, attendu /chat/completions")

    def test_disables_thinking_by_default(self):
        with mock.patch("hrm_sapient.requests.post",
                        return_value=fake_response(payload=chat_payload("ok"))) as post:
            self.client.generate("prompt")

        body = post.call_args[1]["json"]
        self.assertEqual(body.get("chat_template_kwargs"), {"enable_thinking": False})
        self.assertEqual(body["messages"], [{"role": "user", "content": "prompt"}])

    def test_thinking_can_be_re_enabled_per_call(self):
        with mock.patch("hrm_sapient.requests.post",
                        return_value=fake_response(payload=chat_payload("ok"))) as post:
            self.client.generate("prompt", enable_thinking=True)

        body = post.call_args[1]["json"]
        self.assertNotIn("chat_template_kwargs", body,
                         "le raisonnement doit rester actif quand il est demande")

    def test_returns_content_not_reasoning(self):
        payload = chat_payload("la reponse", reasoning="une longue reflexion interne")
        with mock.patch("hrm_sapient.requests.post",
                        return_value=fake_response(payload=payload)):
            self.assertEqual(self.client.generate("prompt"), "la reponse")

    def test_strips_inlined_reasoning_as_safety_net(self):
        # Serveur sans parseur de raisonnement: la reflexion arrive dans content.
        payload = chat_payload('<think>je reflechis</think>{"ok": true}')
        with mock.patch("hrm_sapient.requests.post",
                        return_value=fake_response(payload=payload)):
            self.assertEqual(self.client.generate("prompt"), '{"ok": true}')

    def test_retries_without_kwargs_when_server_rejects_them(self):
        rejected = fake_response(status=400, text="unknown field chat_template_kwargs")
        accepted = fake_response(payload=chat_payload("ok"))

        with mock.patch("hrm_sapient.requests.post",
                        side_effect=[rejected, accepted]) as post:
            result = self.client.generate("prompt")

        self.assertEqual(result, "ok")
        self.assertEqual(post.call_count, 2, "le second essai n'a pas eu lieu")
        self.assertNotIn("chat_template_kwargs", post.call_args_list[1][1]["json"])

    def test_returns_empty_string_on_server_error(self):
        with mock.patch("hrm_sapient.requests.post",
                        return_value=fake_response(status=500, text="boom")):
            self.assertEqual(self.client.generate("prompt"), "")

    def test_returns_empty_string_on_empty_choices(self):
        with mock.patch("hrm_sapient.requests.post",
                        return_value=fake_response(payload={"choices": []})):
            self.assertEqual(self.client.generate("prompt"), "")

    def test_returns_empty_string_on_connection_error(self):
        with mock.patch("hrm_sapient.requests.post",
                        side_effect=OSError("reseau injoignable")):
            self.assertEqual(self.client.generate("prompt"), "")

    def test_handles_null_content(self):
        # vLLM renvoie content=null quand la generation est coupee pendant la reflexion.
        with mock.patch("hrm_sapient.requests.post",
                        return_value=fake_response(payload=chat_payload(None))):
            self.assertEqual(self.client.generate("prompt"), "")

    def test_is_available(self):
        with mock.patch("hrm_sapient.requests.get",
                        return_value=fake_response(status=200)):
            self.assertTrue(self.client.is_available())

        with mock.patch("hrm_sapient.requests.get", side_effect=OSError("down")):
            self.assertFalse(self.client.is_available())


if __name__ == "__main__":
    unittest.main()
