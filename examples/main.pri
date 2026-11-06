struct Person {
    age: int32;
    pub name: string;

    fn new(age int32, name string): Person {
        return Person{
            age,
            name,
        };
    }

    impl {
        fn getAge(): int32 { return self.age; }
    }
}

fn fib(n int): int {
    if (n < 2) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

tobias := Person.new(40, "Tobias");

mut total := 0;
for (i := 0; i < 10; i = i + 1) {
    total = total + fib(i);
}

total + int(tobias.getAge());
