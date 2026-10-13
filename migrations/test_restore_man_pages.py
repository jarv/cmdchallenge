import json
from pathlib import Path
import sqlite3
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parent.parent
NOTICE = (
    "This system has been minimized by removing packages and content that are\n"
    "not required on a system that users do not log into.\n\n"
    "To restore this content, including manpages, you can run the 'unminimize'\n"
    "command. You will still need to ensure the 'man-db' package is installed.\n"
)


class RestoreManPagesTest(unittest.TestCase):
    def setUp(self):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        self.db_path = Path(directory.name) / "cache with spaces.sqlite3"
        self.rows = [
            ("man grep", "hello_world", 5, 0, NOTICE, 11),
            ("man grep", "last_lines", 5, 0, NOTICE, 12),
            ("man grep | tail", "hello_world", 5, 1, NOTICE, 13),
            ("man bash", "hello_world", 5, 0, "BASH(1)\nNAME\n", 14),
            ("echo hello world", "hello_world", 5, 1, "hello world\n", 15),
            ("false", "hello_world", 5, 0, None, 16),
            ("cat notice", "hello_world", 5, 0, NOTICE + "extra\n", 17),
            ("printf notice", "hello_world", 5, 0, NOTICE.rstrip(), 18),
        ]
        with sqlite3.connect(self.db_path) as db:
            db.execute("""CREATE TABLE challenges (
                cmd TEXT, slug TEXT, version INTEGER, correct INTEGER,
                output TEXT, count INTEGER, PRIMARY KEY (cmd, slug, version)
            )""")
            db.executemany("INSERT INTO challenges VALUES (?, ?, ?, ?, ?, ?)", self.rows)

    def contents(self):
        with sqlite3.connect(self.db_path) as db:
            return db.execute("SELECT * FROM challenges ORDER BY cmd, slug, version").fetchall()

    def run_script(self, answer, path=None):
        return subprocess.run(
            ["bash", str(ROOT / "bin/refresh-man-cache"), str(path or self.db_path)],
            input=answer, text=True, capture_output=True,
        )

    def assert_preserved_results(self):
        self.assertEqual(self.contents(), sorted(self.rows[2:], key=lambda row: row[:3]))

    def test_preview_cancel_and_eof_preserve_every_entry(self):
        before = self.contents()
        for answer in ("no\n", ""):
            with self.subTest(answer=answer):
                result = self.run_script(answer)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn("Matching unsuccessful cache entries: 2", result.stdout)
                self.assertIn("Cancelled; no entries deleted.", result.stdout)
                preview = [json.loads(line) for line in result.stdout.splitlines() if line.startswith("{")]
                self.assertEqual({row["slug"] for row in preview}, {"hello_world", "last_lines"})
                self.assertEqual({row["submissions"] for row in preview}, {11, 12})
                self.assertEqual(self.contents(), before)

    def test_confirmed_migration_is_precise_and_idempotent(self):
        result = self.run_script("delete\n")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("Deleted 2 cache entries.", result.stdout)
        self.assert_preserved_results()
        result = self.run_script("")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("Matching unsuccessful cache entries: 0", result.stdout)
        self.assertNotIn("Type delete", result.stdout)
        self.assert_preserved_results()

    def test_sql_can_be_run_directly(self):
        with sqlite3.connect(self.db_path) as db:
            db.executescript((ROOT / "migrations/restore-man-pages.sql").read_text())
        self.assert_preserved_results()

    def test_missing_database_is_not_created(self):
        path = self.db_path.parent / "missing.sqlite3"
        result = self.run_script("delete\n", path)
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(path.exists())


if __name__ == "__main__":
    unittest.main()
