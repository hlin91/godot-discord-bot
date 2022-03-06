# Two Sum

## The Problem
Given an array of integers **nums** and an integer **target**, return *indices of the two numbers such that they add up to **target***.

You may assume that each input would have **exactly one solution**, and you may not use the *same* element twice.

You can return the answer in any order.

## Hint
A really brute force way would be to search for all possible pairs of numbers but that would be too slow. Again, it's best to try out brute force solutions for just for completeness. It is from these brute force solutions that you can come up with optimizations.

So, if we fix one of the numbers, say **x**, we have to scan the entire array to find the next number,**y**, which is **value - x** where value is the input parameter. Can we change our array somehow so that this search becomes faster?

The second train of thought is, without changing the array, can we use additional space somehow? Like maybe a hash map to speed up the search?

## The Solution
```python
def twoSum(self, nums: List[int], target: int) -> List[int]:
        # One pass hash table solution
        h = {}
        for i, n in enumerate(nums):
            t = target - nums[i]
            if t in h:
                return [h[t], i]
            h[n] = i
        return
```

## Explanation
The optimal solution is to iterate through the array and map every element in the array to its index.

We create a hashmap **h** and define **h[n] = i**, such that **nums[i] = n**.

For every **n** in **nums**, we are looking for the number **target - n** to complete the *two-sum*. Define **t = target - n**. If we've encountered **t** before, then it will be in our hashmap, and so we retrieve its *index* and return our solution.

## Complexity Analysis
The *runtime* for this solution is **O(n)**. Every element in **nums** is processed at most once, and retrieving the index of **t** from our hashmap is a **O(1)** operation.

The *space complexity* for this solution is **O(n)**. The hashmap will contain at most all **n** elements in **nums**.


tags: `Array` `Hash Table`
