f = open("test1.txt", "a")
s = "This is a test file to test word count and see how many words it can count"
MB = 1 << 20
size = 2*MB
s = s*size
f.write(s)
f.close()
