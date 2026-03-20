"""Python常⽤库：numpy
（基础运算）、pandas
（简单数据处理）、
scikit-learn（简单模型调
⽤，不搞训练）
"""
import numpy as np

a = np.array([2,3,4],dtype=np.float32)
print(a.dtype)

b = np.zeros((3,4))
print(b)

c = np.ones((3,4),dtype=np.int16)
print(c)

d = np.empty((3,4))
print(d)

a1 = np.array([1,2,3],dtype=np.int32)
a2 = np.array([4,5,6],dtype=np.int32)

print(a1 + a2)

e = np.arange(12).reshape((3,4))
print(e)

f = np.linspace(1,10,12).reshape((3,4))
print(f)

g = np.arange(12).reshape((3,4))
print(g)
print(np.diff(g))
print(np.nonzero(g))

A = np.arange(3,15).reshape((3,4))
print(A)
print(A[2][1])

for column in A.T:
    print(column)

for item in A.flat:
    print(item)

print(A.flatten())

print(np.vstack((a1,a2)))
print(np.hstack((a1,a2)))
