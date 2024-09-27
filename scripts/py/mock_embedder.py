import random
import time
from typing import List
import contextlib

def new_normalized_float32():
    try:
        # Generate a random integer in the range [0, 2^24)
        max_value = 1 << 24
        n = random.randint(0, max_value - 1)

        # Normalize it to the range [-1, 1]
        normalized_float = 2.0 * (float(n) / max_value) - 1.0

        return normalized_float
    except Exception as e:
        raise Exception("Failed to normalize float32") from e

def new_normalized_vector(n):
    vector = [0.0] * n
    for i in range(n):
        vector[i] = new_normalized_float32()

    return vector

def dot_product(v1, v2):
    sum = 0.0
    for i in range(len(v1)):
        sum += v1[i] * v2[i]
    return sum

# Use Gram-Schmidt to return a vector orthogonal to the basis,
# assuming the vectors in the basis are linearly independent.
def new_orthogonal_vector(dim, *basis):
    candidate = new_normalized_vector(dim)

    for b in basis:
        dp = dot_product(candidate, b)
        basis_norm = dot_product(b, b)
        
        for i in range(len(candidate)):
            candidate[i] -= (dp / basis_norm) * b[i]

    return candidate

# Make n linearly independent vectors of size dim.
def new_linearly_independent_vectors(n, dim):
    vectors = []

    for _ in range(n):
        v = new_orthogonal_vector(dim, *vectors)
        vectors.append(v)

    return vectors

def new_score_vector(S, qvector, basis):
    sum_value = 0.0

    # Populate v2 up to dim-1.
    for i in range(len(qvector) - 1):
        sum_value += qvector[i] * basis[i]

    # Calculate v_{2, dim} such that v1 * v2 = 2S - 1:
    basis[-1] = (2 * S - 1 - sum_value) / qvector[-1]

    # If the vectors are not linearly independent, regenerate the dim-1 elements of v2.
    if not linearly_independent(qvector, basis):
        return new_score_vector(S, qvector, basis)

    return basis

# Example usage
dim = 3

doc1 = {"PageContent": "Gabriel García Márquez", "Score": 0.80}
doc2 = {"PageContent": "Gabriela Mistral", "Score": 0.67}
doc3 = {"PageContent": "Miguel de Cervantes", "Score": 0.09}

#query_vector = new_normalized_vector(dim)
#print(query_vector)

query_vector = [-0.5,0.24, 0.71]

marquez = new_orthogonal_vector(dim, query_vector)
print(marquez)

# Create linearly independent vectors for each document
#lin_indep_vectors = new_linearly_independent_vectors(3, dim)

