import sequare_model
import unittest

class TestSequare(unittest.TestCase):

    def test_sequare(self):
        sq = sequare_model.SquareModel(2,3)
        self.assertEqual(sq.pow(), 8)