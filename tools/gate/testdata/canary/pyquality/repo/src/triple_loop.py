def cube(size: int) -> int:
    total = 0
    for x in range(size):
        for y in range(size):
            for z in range(size):
                total += x * y * z
    return total
